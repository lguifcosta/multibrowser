export namespace main {
	
	export class ProfileInfo {
	    profile_dir: string;
	    disk_usage_mb: number;
	    crash_count: number;
	
	    static createFrom(source: any = {}) {
	        return new ProfileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profile_dir = source["profile_dir"];
	        this.disk_usage_mb = source["disk_usage_mb"];
	        this.crash_count = source["crash_count"];
	    }
	}

}

export namespace process {
	
	export class Browser {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new Browser(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class ProcessTelemetry {
	    pid: number;
	    cpu_percent: number;
	    ram_mb: number;
	    lifetime_sec: number;
	    is_running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProcessTelemetry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.cpu_percent = source["cpu_percent"];
	        this.ram_mb = source["ram_mb"];
	        this.lifetime_sec = source["lifetime_sec"];
	        this.is_running = source["is_running"];
	    }
	}

}

export namespace profile {
	
	export class ProfileFlags {
	    restore_last_session: boolean;
	    user_agent: string;
	    lang: string;
	    window_size: string;
	    proxy_server: string;
	    proxy_bypass_list: string;
	    disable_background_networking: boolean;
	    disable_background_timer_throttling: boolean;
	    disable_renderer_backgrounding: boolean;
	    disable_features_translate_ui: boolean;
	    disable_extensions: boolean;
	    disable_sync: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileFlags(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.restore_last_session = source["restore_last_session"];
	        this.user_agent = source["user_agent"];
	        this.lang = source["lang"];
	        this.window_size = source["window_size"];
	        this.proxy_server = source["proxy_server"];
	        this.proxy_bypass_list = source["proxy_bypass_list"];
	        this.disable_background_networking = source["disable_background_networking"];
	        this.disable_background_timer_throttling = source["disable_background_timer_throttling"];
	        this.disable_renderer_backgrounding = source["disable_renderer_backgrounding"];
	        this.disable_features_translate_ui = source["disable_features_translate_ui"];
	        this.disable_extensions = source["disable_extensions"];
	        this.disable_sync = source["disable_sync"];
	    }
	}
	export class Profile {
	    id: string;
	    name: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    last_used: any;
	    status: string;
	    flags: ProfileFlags;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.last_used = this.convertValues(source["last_used"], null);
	        this.status = source["status"];
	        this.flags = this.convertValues(source["flags"], ProfileFlags);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

