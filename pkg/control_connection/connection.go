package controlconnection

type ControlConnectServer interface {
	GetClientToken() (string, error)
}
