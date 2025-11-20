import { BrowserRouter as Router, Navigate, Route, Routes } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ToastProvider } from './contexts/ToastContext';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';
import DashboardPage from './pages/DashboardPage';
import CreateEmployeePage from './pages/CreateEmployeePage';
import CreateProjectPage from './pages/CreateProjectPage';
import EmployeesHubPage from './pages/EmployeesHubPage';
import ProjectsHubPage from './pages/ProjectsHubPage';
import LoginForm from './pages/LoginPage';

function App() {
  return (
    <AuthProvider>
      <ToastProvider>
        <Router>
          <Routes>
            <Route path="/login" element={<LoginForm />} />
            
            <Route element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }>
              <Route path="/" element={<DashboardPage />} />
              
              <Route path="/admin/employees/new" element={
                <ProtectedRoute allowedRoles={['admin', 'director']}>
                  <CreateEmployeePage />
                </ProtectedRoute>
              } />
              
              <Route path="/admin/employees" element={<EmployeesHubPage />} />
              
              <Route path="/projects" element={<ProjectsHubPage />} />
              
              <Route path="/projects/new" element={
                <ProtectedRoute allowedRoles={['manager', 'director']}>
                  <CreateProjectPage />
                </ProtectedRoute>
              } />
            </Route>
            
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Router>
      </ToastProvider>
    </AuthProvider>
  );
}

export default App;
