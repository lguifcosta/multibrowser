/*
 * webkit_exec_fix — LD_PRELOAD shim for cross-distro AppImage portability.
 *
 * Distro builds of WebKitGTK hardcode the absolute path of their helper
 * processes (WebKitNetworkProcess/WebKitWebProcess/WebKitGPUProcess) at
 * compile time (PKGLIBEXECDIR) and ignore WEBKIT_EXEC_PATH in production
 * builds. So an AppImage built on Ubuntu tries to spawn
 *   /usr/lib/x86_64-linux-gnu/webkit2gtk-4.1/WebKitNetworkProcess
 * which does not exist on Arch/Fedora/etc., and the app aborts.
 *
 * This shim intercepts process spawning and rewrites that hardcoded prefix
 * to the helpers bundled inside the AppImage.
 *
 * Env (set by AppRun):
 *   MB_WEBKIT_SYS  prefix baked into libwebkit on the build host
 *                  (default: the Ubuntu multiarch path used by CI)
 *   MB_WEBKIT_DIR  bundled helper dir to redirect to (absolute, runtime)
 */
#define _GNU_SOURCE
#include <string.h>
#include <stdlib.h>
#include <dlfcn.h>
#include <spawn.h>
#include <unistd.h>

static const char *sys_prefix(void) {
    const char *p = getenv("MB_WEBKIT_SYS");
    return (p && *p) ? p : "/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1/";
}

/* Returns a malloc'd rewritten path (leaked on purpose: the caller execs
   immediately after) or NULL when no rewrite applies. */
static const char *rewrite(const char *path) {
    if (!path) return NULL;
    const char *pre = sys_prefix();
    size_t pl = strlen(pre);
    if (strncmp(path, pre, pl) != 0) return NULL;
    const char *base = getenv("MB_WEBKIT_DIR");
    if (!base || !*base) return NULL;
    size_t bl = strlen(base);
    while (bl > 0 && base[bl - 1] == '/') bl--;
    const char *tail = path + pl;
    char *out = malloc(bl + 1 + strlen(tail) + 1);
    if (!out) return NULL;
    memcpy(out, base, bl);
    out[bl] = '/';
    strcpy(out + bl + 1, tail);
    return out;
}

int posix_spawn(pid_t *pid, const char *path,
                const posix_spawn_file_actions_t *fa,
                const posix_spawnattr_t *attr,
                char *const argv[], char *const envp[]) {
    static int (*real)(pid_t *, const char *,
                       const posix_spawn_file_actions_t *,
                       const posix_spawnattr_t *, char *const[], char *const[]);
    if (!real) real = dlsym(RTLD_NEXT, "posix_spawn");
    const char *r = rewrite(path);
    return real(pid, r ? r : path, fa, attr, argv, envp);
}

int execve(const char *path, char *const argv[], char *const envp[]) {
    static int (*real)(const char *, char *const[], char *const[]);
    if (!real) real = dlsym(RTLD_NEXT, "execve");
    const char *r = rewrite(path);
    return real(r ? r : path, argv, envp);
}

int execv(const char *path, char *const argv[]) {
    static int (*real)(const char *, char *const[]);
    if (!real) real = dlsym(RTLD_NEXT, "execv");
    const char *r = rewrite(path);
    return real(r ? r : path, argv);
}
