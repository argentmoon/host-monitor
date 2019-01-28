package report

type Reporter interface {
	Report(string) error
}
