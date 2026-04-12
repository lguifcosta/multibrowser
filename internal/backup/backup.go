package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"multibrowser/internal/lock"
	"multibrowser/internal/logger"
	"multibrowser/internal/profile"
)

var cacheDirs = map[string]bool{
	"Cache":      true,
	"Code Cache": true,
	"GPUCache":   true,
	"Crashpad":   true,
}

type Service struct {
	profileManager *profile.Manager
	lockManager    *lock.Manager
}

func NewService(pm *profile.Manager, lm *lock.Manager) *Service {
	return &Service{
		profileManager: pm,
		lockManager:    lm,
	}
}

func (s *Service) Export(profileID string, excludeCache bool, password string) (string, error) {
	p, err := s.profileManager.Get(profileID)
	if err != nil {
		return "", err
	}

	profileDir := s.profileManager.ProfileDir(p.ID)

	if s.lockManager.IsLocked(profileDir) {
		return "", fmt.Errorf("cannot backup running profile: %s", p.Name)
	}

	backupName := fmt.Sprintf("%s.tar.gz", p.ID)
	if password != "" {
		backupName = fmt.Sprintf("%s.tar.gz.enc", p.ID)
	}
	backupPath := filepath.Join(s.profileManager.BackupsDir(), backupName)

	if password != "" {
		err = s.exportEncrypted(profileDir, backupPath, excludeCache, password)
	} else {
		err = s.exportPlain(profileDir, backupPath, excludeCache)
	}

	if err != nil {
		os.Remove(backupPath)
		return "", err
	}

	if logger.Info != nil {
		logger.Info.Printf("exported backup for profile %s to %s", p.Name, backupPath)
	}

	return backupPath, nil
}

func (s *Service) Import(backupPath, newName, password string) (*profile.Profile, error) {
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup file not found: %s", backupPath)
	}

	p, err := s.profileManager.Create(newName)
	if err != nil {
		return nil, err
	}

	profileDir := s.profileManager.ProfileDir(p.ID)

	if password != "" {
		err = s.importEncrypted(backupPath, profileDir, password)
	} else {
		err = s.importPlain(backupPath, profileDir)
	}

	if err != nil {
		s.profileManager.Delete(p.ID)
		return nil, fmt.Errorf("failed to import backup: %w", err)
	}

	if logger.Info != nil {
		logger.Info.Printf("imported backup from %s as profile %s", backupPath, p.Name)
	}

	return p, nil
}

func (s *Service) exportPlain(srcDir, dstPath string, excludeCache bool) error {
	file, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return writeTarGz(file, srcDir, excludeCache)
}

func (s *Service) exportEncrypted(srcDir, dstPath string, excludeCache bool, password string) error {
	file, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer file.Close()

	key := deriveKey(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// Create tar.gz in memory first, then encrypt
	tmpFile, err := os.CreateTemp("", "multibrowser-backup-*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := writeTarGz(tmpFile, srcDir, excludeCache); err != nil {
		return err
	}

	tmpFile.Seek(0, 0)
	plaintext, err := io.ReadAll(tmpFile)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	_, err = file.Write(ciphertext)
	return err
}

func (s *Service) importPlain(backupPath, dstDir string) error {
	file, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return extractTarGz(file, dstDir)
}

func (s *Service) importEncrypted(backupPath, dstDir, password string) error {
	ciphertext, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	key := deriveKey(password)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("backup file is corrupted")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed (wrong password?): %w", err)
	}

	tmpFile, err := os.CreateTemp("", "multibrowser-import-*.tar.gz")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(plaintext); err != nil {
		return err
	}
	tmpFile.Seek(0, 0)

	return extractTarGz(tmpFile, dstDir)
}

func writeTarGz(w io.Writer, srcDir string, excludeCache bool) error {
	gw := gzip.NewWriter(w)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		// Skip .lock files
		if info.Name() == ".lock" {
			return nil
		}

		// Skip cache directories if requested
		if excludeCache && info.IsDir() {
			parts := strings.Split(relPath, string(filepath.Separator))
			for _, part := range parts {
				if cacheDirs[part] {
					return filepath.SkipDir
				}
			}
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

func extractTarGz(r io.Reader, dstDir string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dstDir, header.Name)

		// Prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dstDir)) {
			return fmt.Errorf("invalid path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}

	return nil
}

func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}
