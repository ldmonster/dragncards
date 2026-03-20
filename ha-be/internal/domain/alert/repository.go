package alert

type AlertRepository interface {
	Create(alert *Alert) error
	List() ([]*Alert, error)
}
