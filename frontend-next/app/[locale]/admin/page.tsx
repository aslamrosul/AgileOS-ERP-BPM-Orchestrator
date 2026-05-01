'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getUser } from '@/lib/auth';

export default function AdminPage() {
  const router = useRouter();
  const user = getUser();

  useEffect(() => {
    if (!user) {
      router.push('/en/login');
    }
  }, [user, router]);

  if (!user) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Redirecting to login...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Admin Panel</h1>
          <p className="text-gray-600 mt-2">System administration and user management</p>
        </div>

        <div className="bg-white p-12 rounded-lg shadow text-center">
          <div className="text-6xl mb-4">⚙️</div>
          <h2 className="text-2xl font-semibold text-gray-900 mb-4">Administration Dashboard</h2>
          <p className="text-gray-600 mb-6">
            Welcome, {user.full_name || user.username}!
          </p>
          <p className="text-gray-500">
            Admin features are available. Manage users, roles, and system settings.
          </p>
          <div className="mt-8 flex gap-4 justify-center">
            <a href="/en" className="px-6 py-3 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300">
              Back to Home
            </a>
            <a href="/en/admin/users" className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
              Manage Users
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
