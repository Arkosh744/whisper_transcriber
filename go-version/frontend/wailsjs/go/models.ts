export namespace main {
	
	export class FileItem {
	    id: string;
	    path: string;
	    name: string;
	    sizeMb: number;
	    status: string;
	    progress: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new FileItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.name = source["name"];
	        this.sizeMb = source["sizeMb"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.error = source["error"];
	    }
	}
	export class LangOption {
	    code: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new LangOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	    }
	}
	export class TranscriptionConfig {
	    language: string;
	    outputFormat: string;
	
	    static createFrom(source: any = {}) {
	        return new TranscriptionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.outputFormat = source["outputFormat"];
	    }
	}

}

