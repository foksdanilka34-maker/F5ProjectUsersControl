package client

type Manager struct {
	Employee *EmployeeClient
	Project  *ProjectClient
}

func NewManager(employeeAddr, projectAddr string) (*Manager, error) {
	empClient, err := NewEmployeeClient(employeeAddr)
	if err != nil {
		return nil, err
	}

	projClient, err := NewProjectClient(projectAddr)
	if err != nil {
		empClient.Close()
		return nil, err
	}

	return &Manager{
		Employee: empClient,
		Project:  projClient,
	}, nil
}

func (m *Manager) Close() error {
	if m.Employee != nil {
		m.Employee.Close()
	}
	if m.Project != nil {
		m.Project.Close()
	}
	return nil
}
