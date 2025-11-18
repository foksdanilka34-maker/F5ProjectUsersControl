import { BrowserRouter as Router, Navigate, Route, Routes } from 'react-router-dom';
import DashboardPage from './pages/DashboardPage';
import CreateEmployeePage from './pages/CreateEmployeePage';
import CreateProjectPage from './pages/CreateProjectPage';
import EmployeesHubPage from './pages/EmployeesHubPage';
import ProjectsHubPage from './pages/ProjectsHubPage';
import LoginForm from './pages/LoginPage';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/login" element={<LoginForm />} />
        <Route path="/admin/employees/new" element={<CreateEmployeePage />} />
        <Route path="/admin/employees" element={<EmployeesHubPage />} />
        <Route path="/projects" element={<ProjectsHubPage />} />
        <Route path="/projects/new" element={<CreateProjectPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Router>
  );
}

export default App;
