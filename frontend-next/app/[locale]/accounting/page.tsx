'use client';

import { useEffect, useState } from 'react';
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Users, 
  FileText, 
  AlertCircle,
  ArrowUpRight,
  ArrowDownRight,
  Calendar,
  RefreshCw
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface DashboardStats {
  totalAssets: number;
  totalLiabilities: number;
  totalEquity: number;
  totalRevenue: number;
  totalExpenses: number;
  netProfit: number;
  accountsCount: number;
  pendingJournals: number;
  overdueInvoices: number;
}

export default function AccountingDashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date());

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      
      // TODO: Replace with actual API call
      // Simulated data for now
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setStats({
        totalAssets: 230000000,
        totalLiabilities: 40000000,
        totalEquity: 200000000,
        totalRevenue: 150000000,
        totalExpenses: 85000000,
        netProfit: 65000000,
        accountsCount: 45,
        pendingJournals: 12,
        overdueInvoices: 3
      });
      
      setLastUpdated(new Date());
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      toast.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  };

  const formatPercentage = (value: number, total: number) => {
    return ((value / total) * 100).toFixed(1);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <p className="text-gray-600">Failed to load dashboard data</p>
          <button
            onClick={fetchDashboardData}
            className="mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const profitMargin = (stats.netProfit / stats.totalRevenue) * 100;

  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Accounting Dashboard</h1>
            <p className="text-gray-600 mt-2">Financial overview and key metrics</p>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-sm text-gray-600">
              <Calendar className="w-4 h-4" />
              <span>Last updated: {lastUpdated.toLocaleTimeString()}</span>
            </div>
            <button
              onClick={fetchDashboardData}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <RefreshCw className="w-4 h-4" />
              <span>Refresh</span>
            </button>
          </div>
        </div>
      </div>

      {/* Key Metrics - Balance Sheet */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {/* Assets */}
        <div className="bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl shadow-sm p-6 border border-blue-200">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-blue-600 rounded-lg flex items-center justify-center">
              <TrendingUp className="w-6 h-6 text-white" />
            </div>
            <span className="text-xs font-medium text-blue-700 bg-blue-200 px-2 py-1 rounded-full">
              Assets
            </span>
          </div>
          <h3 className="text-sm font-medium text-blue-900 mb-1">Total Assets</h3>
          <p className="text-2xl font-bold text-blue-900">{formatCurrency(stats.totalAssets)}</p>
          <div className="mt-4 flex items-center gap-2">
            <ArrowUpRight className="w-4 h-4 text-blue-600" />
            <span className="text-sm text-blue-700">
              {formatPercentage(stats.totalAssets, stats.totalAssets + stats.totalLiabilities)}% of total
            </span>
          </div>
        </div>

        {/* Liabilities */}
        <div className="bg-gradient-to-br from-red-50 to-red-100 rounded-xl shadow-sm p-6 border border-red-200">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-red-600 rounded-lg flex items-center justify-center">
              <TrendingDown className="w-6 h-6 text-white" />
            </div>
            <span className="text-xs font-medium text-red-700 bg-red-200 px-2 py-1 rounded-full">
              Liabilities
            </span>
          </div>
          <h3 className="text-sm font-medium text-red-900 mb-1">Total Liabilities</h3>
          <p className="text-2xl font-bold text-red-900">{formatCurrency(stats.totalLiabilities)}</p>
          <div className="mt-4 flex items-center gap-2">
            <ArrowDownRight className="w-4 h-4 text-red-600" />
            <span className="text-sm text-red-700">
              {formatPercentage(stats.totalLiabilities, stats.totalAssets)}% of assets
            </span>
          </div>
        </div>

        {/* Equity */}
        <div className="bg-gradient-to-br from-emerald-50 to-emerald-100 rounded-xl shadow-sm p-6 border border-emerald-200">
          <div className="flex items-center justify-between mb-4">
            <div className="w-12 h-12 bg-emerald-600 rounded-lg flex items-center justify-center">
              <DollarSign className="w-6 h-6 text-white" />
            </div>
            <span className="text-xs font-medium text-emerald-700 bg-emerald-200 px-2 py-1 rounded-full">
              Equity
            </span>
          </div>
          <h3 className="text-sm font-medium text-emerald-900 mb-1">Total Equity</h3>
          <p className="text-2xl font-bold text-emerald-900">{formatCurrency(stats.totalEquity)}</p>
          <div className="mt-4 flex items-center gap-2">
            <ArrowUpRight className="w-4 h-4 text-emerald-600" />
            <span className="text-sm text-emerald-700">
              {formatPercentage(stats.totalEquity, stats.totalAssets)}% of assets
            </span>
          </div>
        </div>
      </div>

      {/* Profit & Loss */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center">
              <TrendingUp className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-xs text-gray-600">Revenue</p>
              <p className="text-lg font-bold text-gray-900">{formatCurrency(stats.totalRevenue)}</p>
            </div>
          </div>
          <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
            <div className="h-full bg-green-500" style={{ width: '100%' }}></div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 bg-red-100 rounded-lg flex items-center justify-center">
              <TrendingDown className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-xs text-gray-600">Expenses</p>
              <p className="text-lg font-bold text-gray-900">{formatCurrency(stats.totalExpenses)}</p>
            </div>
          </div>
          <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
            <div 
              className="h-full bg-red-500" 
              style={{ width: `${formatPercentage(stats.totalExpenses, stats.totalRevenue)}%` }}
            ></div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 bg-emerald-100 rounded-lg flex items-center justify-center">
              <DollarSign className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-xs text-gray-600">Net Profit</p>
              <p className="text-lg font-bold text-gray-900">{formatCurrency(stats.netProfit)}</p>
            </div>
          </div>
          <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
            <div 
              className="h-full bg-emerald-500" 
              style={{ width: `${profitMargin}%` }}
            ></div>
          </div>
        </div>

        <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-200">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
              <TrendingUp className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-xs text-gray-600">Profit Margin</p>
              <p className="text-lg font-bold text-gray-900">{profitMargin.toFixed(1)}%</p>
            </div>
          </div>
          <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
            <div 
              className="h-full bg-purple-500" 
              style={{ width: `${profitMargin}%` }}
            ></div>
          </div>
        </div>
      </div>

      {/* Quick Actions & Alerts */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Quick Actions */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-bold text-gray-900 mb-4">Quick Actions</h2>
          <div className="space-y-3">
            <Link
              href="/accounting/chart-of-accounts"
              className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors group"
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-emerald-100 rounded-lg flex items-center justify-center">
                  <FileText className="w-5 h-5 text-emerald-600" />
                </div>
                <div>
                  <p className="font-medium text-gray-900">Chart of Accounts</p>
                  <p className="text-sm text-gray-600">{stats.accountsCount} accounts</p>
                </div>
              </div>
              <ArrowUpRight className="w-5 h-5 text-gray-400 group-hover:text-emerald-600" />
            </Link>

            <Link
              href="/accounting/journal-entries"
              className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors group"
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
                  <FileText className="w-5 h-5 text-blue-600" />
                </div>
                <div>
                  <p className="font-medium text-gray-900">Journal Entries</p>
                  <p className="text-sm text-gray-600">{stats.pendingJournals} pending</p>
                </div>
              </div>
              <ArrowUpRight className="w-5 h-5 text-gray-400 group-hover:text-blue-600" />
            </Link>

            <Link
              href="/accounting/reports"
              className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors group"
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center">
                  <FileText className="w-5 h-5 text-purple-600" />
                </div>
                <div>
                  <p className="font-medium text-gray-900">Financial Reports</p>
                  <p className="text-sm text-gray-600">View all reports</p>
                </div>
              </div>
              <ArrowUpRight className="w-5 h-5 text-gray-400 group-hover:text-purple-600" />
            </Link>
          </div>
        </div>

        {/* Alerts & Notifications */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-bold text-gray-900 mb-4">Alerts & Notifications</h2>
          <div className="space-y-3">
            {stats.overdueInvoices > 0 && (
              <div className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg">
                <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-red-900">Overdue Invoices</p>
                  <p className="text-sm text-red-700">
                    {stats.overdueInvoices} invoice{stats.overdueInvoices > 1 ? 's' : ''} overdue
                  </p>
                  <Link
                    href="/accounting/invoices?status=overdue"
                    className="text-sm text-red-600 hover:text-red-800 font-medium mt-1 inline-block"
                  >
                    View details →
                  </Link>
                </div>
              </div>
            )}

            {stats.pendingJournals > 0 && (
              <div className="flex items-start gap-3 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                <AlertCircle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-yellow-900">Pending Journal Entries</p>
                  <p className="text-sm text-yellow-700">
                    {stats.pendingJournals} journal entries awaiting posting
                  </p>
                  <Link
                    href="/accounting/journal-entries?status=draft"
                    className="text-sm text-yellow-600 hover:text-yellow-800 font-medium mt-1 inline-block"
                  >
                    Review now →
                  </Link>
                </div>
              </div>
            )}

            <div className="flex items-start gap-3 p-4 bg-emerald-50 border border-emerald-200 rounded-lg">
              <TrendingUp className="w-5 h-5 text-emerald-600 flex-shrink-0 mt-0.5" />
              <div>
                <p className="font-medium text-emerald-900">Healthy Profit Margin</p>
                <p className="text-sm text-emerald-700">
                  Your profit margin is {profitMargin.toFixed(1)}% - above industry average
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
