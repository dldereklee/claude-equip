package profile

type Index struct {
	Profiles []IndexProfile `json:"profiles"`
}

type IndexProfile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Tool        string `json:"tool"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
}

type Manifest struct {
	Name  string       `json:"name"`
	Tool  string       `json:"tool"`
	Files []FileTarget `json:"files"`
}

type FileTarget struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Registry struct {
	Sources  []SourceRef       `json:"sources"`
	Profiles []ProfileRegistry `json:"profiles"`
}

type SourceRef struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	CacheDir string `json:"cache_dir"`
}

type ProfileRegistry struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Tool        string       `json:"tool"`
	Description string       `json:"description,omitempty"`
	SourceID    string       `json:"source_id"`
	ManifestRel string       `json:"manifest_rel"`
	Files       []FileTarget `json:"files"`
}
