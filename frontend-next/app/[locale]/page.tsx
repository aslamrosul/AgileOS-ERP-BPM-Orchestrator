"use client";

import Link from 'next/link';

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-16">
        {/* Header */}
        <div className="text-center mb-16">
          <h1 className="text-6xl font-bold text-gray-900 mb-4">
            AgileOS
          </h1>
          <p className="text-xl text-gray-600 mb-8">
            Enterprise Business Process Management Platform with Real-time Notifications
          </p>
          
          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/en/login"
              className="inline-flex items-center justify-center gap-2 bg-blue-600 text-white px-8 py-4 rounded-lg text-lg font-semibold hover:bg-blue-700 transition-colors"
            >
              Login
            </Link>
            <Link
              href="/en/workflow"
              className="inline-flex items-center justify-center gap-2 bg-indigo-600 text-white px-8 py-4 rounded-lg text-lg font-semibold hover:bg-indigo-700 transition-colors"
            >
              Workflow Builder
            </Link>
            <Link
              href="/en/analytics"
              className="inline-flex items-center justify-center gap-2 bg-purple-600 text-white px-8 py-4 rounded-lg text-lg font-semibold hover:bg-purple-700 transition-colors"
            >
              Analytics Dashboard
            </Link>
          </div>
        </div>

        {/* Features Grid */}
        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          <div className="bg-white p-8 rounded-xl shadow-lg hover:shadow-xl transition-shadow">
            <div className="w-12 h-12 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <svg className="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2">Visual Workflow Builder</h3>
            <p className="text-gray-600">
              Drag-and-drop interface to create complex business workflows without coding
            </p>
          </div>

          <div className="bg-white p-8 rounded-xl shadow-lg hover:shadow-xl transition-shadow">
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2">Enterprise Ready</h3>
            <p className="text-gray-600">
              Built with Go and SurrealDB for high performance and scalability
            </p>
          </div>

          <div className="bg-white p-8 rounded-xl shadow-lg hover:shadow-xl transition-shadow">
            <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
            </div>
            <h3 className="text-xl font-semibold mb-2">Real-time Notifications</h3>
            <p className="text-gray-600">
              WebSocket-powered real-time notifications for task assignments and approvals
            </p>
          </div>
        </div>

        {/* Quick Links */}
        <div className="mt-16 text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Quick Access</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 max-w-4xl mx-auto">
            <Link href="/en/accounting" className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition-shadow">
              <div className="text-3xl mb-2">💰</div>
              <div className="font-semibold">Accounting</div>
              <div className="text-sm text-gray-500">COA, Journals, AP</div>
            </Link>
            <Link href="/en/admin" className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition-shadow">
              <div className="text-3xl mb-2">⚙️</div>
              <div className="font-semibold">Admin</div>
              <div className="text-sm text-gray-500">User Management</div>
            </Link>
            <Link href="/en/audit" className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition-shadow">
              <div className="text-3xl mb-2">📋</div>
              <div className="font-semibold">Audit Trail</div>
              <div className="text-sm text-gray-500">Compliance & Logs</div>
            </Link>
            <Link href="/en/workflow" className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition-shadow">
              <div className="text-3xl mb-2">🔄</div>
              <div className="font-semibold">Workflows</div>
              <div className="text-sm text-gray-500">Process Builder</div>
            </Link>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-16 text-center text-gray-600">
          <p className="text-sm">
            AgileOS v1.0 - Enterprise Business Process Management Platform
          </p>
          <p className="text-xs mt-2">
            Powered by Go, SurrealDB, Next.js, and NATS
          </p>
        </div>
      </div>
    </div>
  );
}
