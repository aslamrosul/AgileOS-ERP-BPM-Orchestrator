'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { getUser, authenticatedFetch } from '@/lib/auth';
import {
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { AlertTriangle, TrendingUp, Clock, CheckCircle, RefreshCw, Activity } from 'lucide-react';

interface AnalyticsData {
  summary: {
    total_processes: number;
    active_processes: number;
    completed_processes: number;
    total_tasks: number;
    completed_tasks: number;
    pending_tasks: number;
    avg_completion_time_hours: number;
    overall_sla_compliance_rate: number;
  };
  department_metrics: Array<{
    department: string;
    total_tasks: number;
    completed_tasks: number;
    pending_tasks: number;
    avg_duration_hours: number;
    sla_violations: number;
    high_latency: boolean;
    recommendation?: string;
  }>;
  task_status_breakdown: Array<{
    status: string;
    count: number;
    percentage: number;
  }>;
  bottlenecks: Array<{
    step_name: string;
    assigned_to: string;
    avg_duration_hours: number;
    total_tasks: number;
    sla_compliance: number;
  }>;
  insights: Array<{
    type: string;
    category: string;
    title: string;
    description: string;
    recommendation: string;
    priority: string;
  }>;
}

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6'];

export default function AnalyticsPage() {
  const router = useRouter();
  const currentUser = getUser();
  
  const [data, setData] = useState<AnalyticsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAnalytics = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/analytics/overview?days=7`
      );

      if (!response.ok) {
        throw new Error('Failed to fetch analytics');
      }

      const analyticsData = await response.json();
      setData(analyticsData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load analytics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!currentUser) {
      router.push('/en/login');
      return;
    }
    
    if (!['admin', 'manager'].includes(currentUser.role)) {
      router.push('/en');
      return;
    }

    fetchAnalytics();
  }, []); // Empty dependency array - only run once on mount

  if (!currentUser || loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <AlertTriangle className="h-12 w-12 text-red-500 mx-auto" />
          <p className="mt-4 text-gray-600">{error || 'No data available'}</p>
          <button
            onClick={fetchAnalytics}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-3">
                <Activity className="w-8 h-8 text-indigo-600" />
                Business Analytics Dashboard
              </h1>
              <p className="text-gray-600 mt-2">Last 7 days performance overview</p>
            </div>
            <button
              onClick={fetchAnalytics}
              className="flex items-center gap-2 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
            >
              <RefreshCw className="w-4 h-4" />
              Refresh
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* KPI Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <KPICard
            title="Total Processes"
            value={data.summary.total_processes}
            icon={<TrendingUp className="h-6 w-6" />}
            color="blue"
          />
          <KPICard
            title="Completed Tasks"
            value={data.summary.completed_tasks}
            icon={<CheckCircle className="h-6 w-6" />}
            color="green"
          />
          <KPICard
            title="Pending Tasks"
            value={data.summary.pending_tasks}
            icon={<Clock className="h-6 w-6" />}
            color="yellow"
          />
          <KPICard
            title="Avg Completion Time"
            value={`${data.summary.avg_completion_time_hours.toFixed(1)}h`}
            icon={<Clock className="h-6 w-6" />}
            color="purple"
          />
        </div>

        {/* Charts Row 1 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Department Efficiency Bar Chart */}
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Department Efficiency</h2>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={data.department_metrics}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="department" />
                <YAxis label={{ value: 'Hours', angle: -90, position: 'insideLeft' }} />
                <Tooltip />
                <Legend />
                <Bar dataKey="avg_duration_hours" fill="#3b82f6" name="Avg Duration" />
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Task Status Pie Chart */}
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Task Status Distribution</h2>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={data.task_status_breakdown}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ status, percentage }) => `${status}: ${percentage.toFixed(1)}%`}
                  outerRadius={100}
                  fill="#8884d8"
                  dataKey="count"
                >
                  {data.task_status_breakdown.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Bottlenecks Section */}
        {data.bottlenecks && data.bottlenecks.length > 0 && (
          <div className="bg-white p-6 rounded-lg shadow mb-8">
            <h2 className="text-xl font-semibold mb-4 flex items-center">
              <AlertTriangle className="h-5 w-5 text-red-500 mr-2" />
              Top Bottlenecks
            </h2>
            <div className="space-y-4">
              {data.bottlenecks.map((bottleneck, index) => (
                <div key={index} className="border-l-4 border-red-500 pl-4 py-2">
                  <h3 className="font-semibold text-gray-900">{bottleneck.step_name}</h3>
                  <p className="text-sm text-gray-600">
                    Assigned to: {bottleneck.assigned_to}
                  </p>
                  <p className="text-sm text-gray-600">
                    Avg Duration: {bottleneck.avg_duration_hours.toFixed(1)} hours
                  </p>
                  <p className="text-sm text-gray-600">
                    SLA Compliance: {bottleneck.sla_compliance.toFixed(1)}%
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Business Insights */}
        {data.insights && data.insights.length > 0 && (
          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Business Insights & Recommendations</h2>
            <div className="space-y-4">
              {data.insights.map((insight, index) => (
                <InsightCard key={index} insight={insight} />
              ))}
            </div>
          </div>
        )}

        {/* Back Button */}
        <div className="mt-8">
          <a
            href="/en"
            className="inline-flex items-center gap-2 px-6 py-3 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
          >
            ← Back to Home
          </a>
        </div>
      </div>
    </div>
  );
}

function KPICard({
  title,
  value,
  icon,
  color,
}: {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  color: string;
}) {
  const colorClasses = {
    blue: 'bg-blue-100 text-blue-600',
    green: 'bg-green-100 text-green-600',
    yellow: 'bg-yellow-100 text-yellow-600',
    purple: 'bg-purple-100 text-purple-600',
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-600">{title}</p>
          <p className="text-2xl font-bold text-gray-900 mt-2">{value}</p>
        </div>
        <div className={`p-3 rounded-full ${colorClasses[color as keyof typeof colorClasses]}`}>
          {icon}
        </div>
      </div>
    </div>
  );
}

function InsightCard({ insight }: { insight: any }) {
  const typeColors = {
    warning: 'border-yellow-500 bg-yellow-50',
    success: 'border-green-500 bg-green-50',
    info: 'border-blue-500 bg-blue-50',
  };

  const priorityBadges = {
    high: 'bg-red-100 text-red-800',
    medium: 'bg-yellow-100 text-yellow-800',
    low: 'bg-green-100 text-green-800',
  };

  return (
    <div className={`border-l-4 p-4 rounded ${typeColors[insight.type as keyof typeof typeColors]}`}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <h3 className="font-semibold text-gray-900">{insight.title}</h3>
            <span className={`px-2 py-1 text-xs rounded ${priorityBadges[insight.priority as keyof typeof priorityBadges]}`}>
              {insight.priority.toUpperCase()}
            </span>
          </div>
          <p className="text-sm text-gray-700 mb-2">{insight.description}</p>
          <p className="text-sm text-gray-600 italic">
            💡 {insight.recommendation}
          </p>
        </div>
      </div>
    </div>
  );
}
