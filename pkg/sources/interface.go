package sources

// Source interface
type Source interface {
	Name() string
	Available(app string) (bool, error)
	Installed(app string) (bool, error)
	Install(app string) error
	Remove(app string) error
	Update() error
	Search(app string) (bool, error)
	InstalledCount() (int, error)
}
