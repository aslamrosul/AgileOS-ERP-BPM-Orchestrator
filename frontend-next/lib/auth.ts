// Authentication utilities for Next.js frontend

const TOKEN_KEY = 'agileos_access_token';
const REFRESH_TOKEN_KEY = 'agileos_refresh_token';
const USER_KEY = 'agileos_user';

export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  full_name: string;
  department?: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

// Token management
export const setTokens = (accessToken: string, refreshToken: string) => {
  if (typeof window !== 'undefined') {
    localStorage.setItem(TOKEN_KEY, accessToken);
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }
};

export const getAccessToken = (): string | null => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(TOKEN_KEY);
  }
  return null;
};

export const getRefreshToken = (): string | null => {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(REFRESH_TOKEN_KEY);
  }
  return null;
};

export const clearTokens = () => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }
};

// User management
export const setUser = (user: User) => {
  if (typeof window !== 'undefined') {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
};

export const getUser = (): User | null => {
  if (typeof window !== 'undefined') {
    const userStr = localStorage.getItem(USER_KEY);
    if (userStr) {
      try {
        return JSON.parse(userStr);
      } catch {
        return null;
      }
    }
  }
  return null;
};

// Authentication status
export const isAuthenticated = (): boolean => {
  return !!getAccessToken();
};

// Logout
export const logout = () => {
  clearTokens();
  if (typeof window !== 'undefined') {
    window.location.href = '/login';
  }
};

// API request with authentication
export const authenticatedFetch = async (
  url: string,
  options: RequestInit = {}
): Promise<Response> => {
  const token = getAccessToken();

  const headers = {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
    ...options.headers,
  };

  let response = await fetch(url, {
    ...options,
    headers,
  });

  // If unauthorized, try to refresh token
  if (response.status === 401) {
    const refreshToken = getRefreshToken();
    if (refreshToken) {
      try {
        const refreshResponse = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/refresh`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken }),
          }
        );

        if (refreshResponse.ok) {
          const data = await refreshResponse.json();
          setTokens(data.access_token, refreshToken);

          // Retry original request with new token
          headers.Authorization = `Bearer ${data.access_token}`;
          response = await fetch(url, {
            ...options,
            headers,
          });
        } else {
          // Refresh failed, logout
          logout();
        }
      } catch (error) {
        console.error('Token refresh failed:', error);
        logout();
      }
    } else {
      // No refresh token, logout
      logout();
    }
  }

  return response;
};

// Login function
export const login = async (
  username: string,
  password: string
): Promise<LoginResponse> => {
  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/login`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Login failed');
  }

  const data: LoginResponse = await response.json();
  setTokens(data.access_token, data.refresh_token);
  setUser(data.user);

  return data;
};

// Register function
export const register = async (userData: {
  username: string;
  email: string;
  password: string;
  full_name: string;
  department?: string;
  role: string;
}): Promise<LoginResponse> => {
  const response = await fetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/register`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(userData),
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Registration failed');
  }

  const data: LoginResponse = await response.json();
  setTokens(data.access_token, data.refresh_token);
  setUser(data.user);

  return data;
};

// Get current user profile
export const getProfile = async (): Promise<User> => {
  const response = await authenticatedFetch(
    `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/profile`
  );

  if (!response.ok) {
    throw new Error('Failed to get profile');
  }

  return response.json();
};

// Role-based access control
export const hasRole = (requiredRole: string | string[]): boolean => {
  const user = getUser();
  if (!user) return false;

  if (Array.isArray(requiredRole)) {
    return requiredRole.includes(user.role);
  }

  return user.role === requiredRole;
};

export const isAdmin = (): boolean => hasRole('admin');
export const isManager = (): boolean => hasRole(['admin', 'manager']);
