'use client';

import { useEffect, useState } from 'react';
import { 
  Plus, Search, Filter, Download, Eye, Edit, Trash2, CheckCircle,
  Clock, XCircle, AlertCircle, TrendingUp, Calendar, DollarSign,
  RefreshCw, BarChart3, PieChart, Activity, Target, Award,
  ArrowUp, ArrowDown, Minus, FileText, Users, Building
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface Budget {
  id: string;
  budget_name: string;
  fiscal_year: number;
  account_id: string;
  account_code: string;
  account_name: string;
  department: string;
  period_type: 'monthly' | 'quarterly' | 'yearly';
  total_amount: number;
  allocated_amount: number;
  spent_amount: number;
  remaining_amount: number;
  status: 'draft' | 'submitted' | 'approved' | 'active' | 'closed';
  start_date: string;
  end_date: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

interface Account {
  id: string;
  account_code: string;
  account_name: string;
  account_type: string;
}

interface BudgetFormData {
  budget_name: string;
  fiscal_year: number;
  account_id: string;
  department: string;
  period_type: 'monthly' | 'quarterly' | 'yearly';
  total_amount: number;
  start_date: string;
  end_date: string;
  status: string;
}

export default function BudgetPage() {
  const [budgets, setBudgets] = useState<Budget[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [filteredBudgets, setFilteredBudgets] = useState<Budget[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [fiscalYearFilter, setFiscalYearFilter] = useState<string>('all');
  const [departmentFilter, setDepartmentFilter] = useState<string>('all');
  const [periodTypeFilter, setPeriodTypeFilter] = useState<string>('all');
  
  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showViewModal, setShowViewModal] = useState(false);
  const [showVarianceModal, setShowVarianceModal] = useState(false);
  const [selectedBudget, setSelectedBudget] = useState<Budget | null>(null);
  const [varianceData, setVarianceData] = useState<any>(null);
  
  // Form state
  const [formData, setFormData] = useState<BudgetFormData>({
    budget_name: '',
    fiscal_year: new Date().getFullYear(),
    account_id: '',
    department: '',
    period_type: 'yearly',
    total_amount: 0,
    start_date: '',
    end_date: '',
    status: 'draft'
  });

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    filterBudgets();
  }, [budgets, searchTerm, statusFilter, fiscalYearFilter, departmentFilter, periodTypeFilter]);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      const [budgetsRes, accountsRes] = await Promise.all([
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/budgets`),
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts`)
      ]);

      if (!budgetsRes.ok || !accountsRes.ok) {
        throw new Error('Failed to fetch data');
      }

      const [budgetsData, accountsData] = await Promise.all([
        budgetsRes.json(),
        accountsRes.json()
      ]);

      setBudgets(budgetsData || []);
      setAccounts(accountsData || []);
    } catch (error) {
      console.error('Failed to fetch data:', error);
      toast.error('Failed to load budgets');
    } finally {
      setLoading(false);
    }
  };

  const filterBudgets = () => {
    let filtered = [...budgets];

    if (searchTerm) {
      filtered = filtered.filter(budget =>
        budget.budget_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        budget.account_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        budget.department.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    if (statusFilter !== 'all') {
      filtered = filtered.filter(budget => budget.status === statusFilter);
    }

    if (fiscalYearFilter !== 'all') {
      filtered = filtered.filter(budget => budget.fiscal_year.toString() === fiscalYearFilter);
    }

    if (departmentFilter !== 'all') {
      filtered = filtered.filter(budget => budget.department === departmentFilter);
    }

    if (periodTypeFilter !== 'all') {
      filtered = filtered.filter(budget => budget.period_type === periodTypeFilter);
    }

    setFilteredBudgets(filtered);
  };

  // CRUD Operations
  const handleCreateBudget = async () => {
    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/budgets`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(formData)
        }
      );

      if (!response.ok) throw new Error('Failed to create budget');

      toast.success('Budget created successfully');
      setShowCreateModal(false);
      resetForm();
      fetchData();
    } catch (error) {
      console.error('Failed to create budget:', error);
      toast.error('Failed to create budget');
    }
  };

  const handleUpdateBudget = async () => {
    if (!selectedBudget) return;

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/budgets/${selectedBudget.id}`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(formData)
        }
      );

      if (!response.ok) throw new Error('Failed to update budget');

      toast.success('Budget updated successfully');
      setShowEditModal(false);
      setSelectedBudget(null);
      resetForm();
      fetchData();
    } catch (error) {
      console.error('Failed to update budget:', error);
      toast.error('Failed to update budget');
    }
  };

  const handleDeleteBudget = async (budgetId: string) => {
    if (!confirm('Are you sure you want to delete this budget? This action cannot be undone.')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/budgets/${budgetId}`,
        { method: 'DELETE' }
      );

      if (!response.ok) throw new Error('Failed to delete budget');

      toast.success('Budget deleted successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to delete budget:', error);
      toast.error('Failed to delete budget');
    }
  };

  const handleApproveBudget = async (budgetId: string) => {
    if (!confirm('Are you sure you want to approve this budget?')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/budgets/${budgetId}/approve`,
        { method: 'POST' }
      );

      if (!response.ok) throw new Error('Failed to approve budget');

      toast.success('Budget approved successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to approve budget:', error);
      toast.error('Failed to approve budget');
    }
  };

  const handleViewVariance = async (budget: Budget) => {
    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/budgets/${budget.id}/variance`
      );

      if (!response.ok) throw new Error('Failed to fetch variance');

      const data = await response.json();
      setVarianceData(data);
      setSelectedBudget(budget);
      setShowVarianceModal(true);
    } catch (error) {
      console.error('Failed to fetch variance:', error);
      toast.error('Failed to fetch variance analysis');
    }
  };

  // Modal handlers
  const openCreateModal = () => {
    resetForm();
    setShowCreateModal(true);
  };

  const openEditModal = (budget: Budget) => {
    setSelectedBudget(budget);
    setFormData({
      budget_name: budget.budget_name,
      fiscal_year: budget.fiscal_year,
      account_id: budget.account_id,
      department: budget.department,
      period_type: budget.period_type,
      total_amount: budget.total_amount,
      start_date: budget.start_date,
      end_date: budget.end_date,
      status: budget.status
    });
    setShowEditModal(true);
  };

  const openViewModal = (budget: Budget) => {
    setSelectedBudget(budget);
    setShowViewModal(true);
  };

  const resetForm = () => {
    setFormData({
      budget_name: '',
      fiscal_year: new Date().getFullYear(),
      account_id: '',
      department: '',
      period_type: 'yearly',
      total_amount: 0,
      start_date: '',
      end_date: '',
      status: 'draft'
    });
  };

  const closeAllModals = () => {
    setShowCreateModal(false);
    setShowEditModal(false);
    setShowViewModal(false);
    setShowVarianceModal(false);
    setSelectedBudget(null);
    setVarianceData(null);
  };

  // Helper functions
  const getStatusBadge = (status: string) => {
    const badges: Record<string, { bg: string; text: string; icon: any }> = {
      draft: { bg: 'bg-gray-100', text: 'text-gray-800', icon: Clock },
      submitted: { bg: 'bg-yellow-100', text: 'text-yellow-800', icon: AlertCircle },
      approved: { bg: 'bg-blue-100', text: 'text-blue-800', icon: CheckCircle },
      active: { bg: 'bg-green-100', text: 'text-green-800', icon: Activity },
      closed: { bg: 'bg-red-100', text: 'text-red-800', icon: XCircle }
    };
    const badge = badges[status] || { bg: 'bg-gray-100', text: 'text-gray-800', icon: FileText };
    const Icon = badge.icon;
    
    return (
      <span className={`inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        <Icon className="w-3 h-3" />
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
  };

  const getPeriodTypeBadge = (type: string) => {
    const badges: Record<string, { bg: string; text: string }> = {
      monthly: { bg: 'bg-purple-100', text: 'text-purple-800' },
      quarterly: { bg: 'bg-blue-100', text: 'text-blue-800' },
      yearly: { bg: 'bg-green-100', text: 'text-green-800' }
    };
    const badge = badges[type] || { bg: 'bg-gray-100', text: 'text-gray-800' };
    
    return (
      <span className={`inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        {type.charAt(0).toUpperCase() + type.slice(1)}
      </span>
    );
  };

  const getVarianceIndicator = (variance: number, percentage: number) => {
    if (variance > 0) {
      return (
        <div className="flex items-center gap-1 text-red-600">
          <ArrowUp className="w-4 h-4" />
          <span className="font-semibold">{percentage.toFixed(1)}% Over</span>
        </div>
      );
    } else if (variance < 0) {
      return (
        <div className="flex items-center gap-1 text-green-600">
          <ArrowDown className="w-4 h-4" />
          <span className="font-semibold">{Math.abs(percentage).toFixed(1)}% Under</span>
        </div>
      );
    } else {
      return (
        <div className="flex items-center gap-1 text-gray-600">
          <Minus className="w-4 h-4" />
          <span className="font-semibold">On Budget</span>
        </div>
      );
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  const calculateUtilization = (spent: number, total: number) => {
    if (total === 0) return 0;
    return (spent / total) * 100;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading budgets...</p>
        </div>
      </div>
    );
  }

  const stats = {
    total: budgets.length,
    draft: budgets.filter(b => b.status === 'draft').length,
    submitted: budgets.filter(b => b.status === 'submitted').length,
    approved: budgets.filter(b => b.status === 'approved').length,
    active: budgets.filter(b => b.status === 'active').length,
    totalBudget: budgets.reduce((sum, b) => sum + b.total_amount, 0),
    totalSpent: budgets.reduce((sum, b) => sum + b.spent_amount, 0),
    totalRemaining: budgets.reduce((sum, b) => sum + b.remaining_amount, 0)
  };

  const utilizationRate = stats.totalBudget > 0 ? (stats.totalSpent / stats.totalBudget) * 100 : 0;

  const fiscalYears = [...new Set(budgets.map(b => b.fiscal_year))].sort((a, b) => b - a);
  const departments = [...new Set(budgets.map(b => b.department))].filter(d => d);

  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Budget Management</h1>
            <p className="text-gray-600 mt-2">Plan and track your budgets</p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => toast.info('Export feature coming soon')}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span>Export</span>
            </button>
            <button
              onClick={openCreateModal}
              className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              <span>New Budget</span>
            </button>
          </div>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Total Budgets</p>
          <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
        </div>
        <div className="bg-gray-50 rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-700 mb-1">Draft</p>
          <p className="text-2xl font-bold text-gray-900">{stats.draft}</p>
        </div>
        <div className="bg-yellow-50 rounded-lg shadow-sm p-4 border border-yellow-200">
          <p className="text-xs text-yellow-700 mb-1">Pending Approval</p>
          <p className="text-2xl font-bold text-yellow-900">{stats.submitted}</p>
        </div>
        <div className="bg-blue-50 rounded-lg shadow-sm p-4 border border-blue-200">
          <p className="text-xs text-blue-700 mb-1">Approved</p>
          <p className="text-2xl font-bold text-blue-900">{stats.approved}</p>
        </div>
        <div className="bg-green-50 rounded-lg shadow-sm p-4 border border-green-200">
          <p className="text-xs text-green-700 mb-1">Active</p>
          <p className="text-2xl font-bold text-green-900">{stats.active}</p>
        </div>
      </div>

      {/* Financial Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Total Budget</p>
            <Target className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalBudget)}</p>
          <p className="text-xs opacity-75 mt-2">Allocated amount</p>
        </div>
        <div className="bg-gradient-to-br from-red-500 to-red-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Total Spent</p>
            <TrendingUp className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalSpent)}</p>
          <p className="text-xs opacity-75 mt-2">{utilizationRate.toFixed(1)}% utilized</p>
        </div>
        <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Remaining</p>
            <Award className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalRemaining)}</p>
          <p className="text-xs opacity-75 mt-2">{(100 - utilizationRate).toFixed(1)}% available</p>
        </div>
        <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Utilization Rate</p>
            <PieChart className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{utilizationRate.toFixed(1)}%</p>
          <p className="text-xs opacity-75 mt-2">Budget usage</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm p-6 border border-gray-200 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-6 gap-4">
          <div className="md:col-span-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search budgets..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>
          </div>

          <div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Status</option>
              <option value="draft">Draft</option>
              <option value="submitted">Submitted</option>
              <option value="approved">Approved</option>
              <option value="active">Active</option>
              <option value="closed">Closed</option>
            </select>
          </div>

          <div>
            <select
              value={fiscalYearFilter}
              onChange={(e) => setFiscalYearFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Years</option>
              {fiscalYears.map(year => (
                <option key={year} value={year}>{year}</option>
              ))}
            </select>
          </div>

          <div>
            <select
              value={periodTypeFilter}
              onChange={(e) => setPeriodTypeFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Periods</option>
              <option value="monthly">Monthly</option>
              <option value="quarterly">Quarterly</option>
              <option value="yearly">Yearly</option>
            </select>
          </div>

          <div>
            <button
              onClick={fetchData}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <RefreshCw className="w-4 h-4" />
              <span>Refresh</span>
            </button>
          </div>
        </div>

        {departments.length > 0 && (
          <div className="mt-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">Department</label>
            <select
              value={departmentFilter}
              onChange={(e) => setDepartmentFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Departments</option>
              {departments.map(dept => (
                <option key={dept} value={dept}>{dept}</option>
              ))}
            </select>
          </div>
        )}
      </div>

      {/* Budgets Table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Budget Name
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Account
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Period
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Budget Amount
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Spent
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Remaining
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredBudgets.map((budget) => {
                const utilization = calculateUtilization(budget.spent_amount, budget.total_amount);
                const variance = budget.spent_amount - budget.total_amount;
                const variancePercentage = budget.total_amount > 0 ? (variance / budget.total_amount) * 100 : 0;
                
                return (
                  <tr key={budget.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Target className="w-4 h-4 text-gray-400" />
                        <div>
                          <p className="text-sm font-medium text-gray-900">{budget.budget_name}</p>
                          <p className="text-xs text-gray-500">FY {budget.fiscal_year}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <p className="text-sm font-medium text-gray-900">{budget.account_name}</p>
                        <p className="text-xs text-gray-500">{budget.account_code}</p>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getPeriodTypeBadge(budget.period_type)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(budget.status)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <p className="text-sm font-medium text-gray-900">{formatCurrency(budget.total_amount)}</p>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <div>
                        <p className="text-sm font-medium text-red-600">{formatCurrency(budget.spent_amount)}</p>
                        <p className="text-xs text-gray-500">{utilization.toFixed(1)}%</p>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <div>
                        <p className="text-sm font-medium text-green-600">{formatCurrency(budget.remaining_amount)}</p>
                        <p className="text-xs text-gray-500">{(100 - utilization).toFixed(1)}%</p>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => openViewModal(budget)}
                          className="p-1.5 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          title="View Details"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleViewVariance(budget)}
                          className="p-1.5 text-gray-600 hover:text-purple-600 hover:bg-purple-50 rounded transition-colors"
                          title="View Variance"
                        >
                          <BarChart3 className="w-4 h-4" />
                        </button>
                        {budget.status === 'draft' && (
                          <>
                            <button
                              onClick={() => openEditModal(budget)}
                              className="p-1.5 text-gray-600 hover:text-emerald-600 hover:bg-emerald-50 rounded transition-colors"
                              title="Edit Budget"
                            >
                              <Edit className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() => handleDeleteBudget(budget.id)}
                              className="p-1.5 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                              title="Delete Budget"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </>
                        )}
                        {budget.status === 'submitted' && (
                          <button
                            onClick={() => handleApproveBudget(budget.id)}
                            className="p-1.5 text-gray-600 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
                            title="Approve Budget"
                          >
                            <CheckCircle className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        {filteredBudgets.length === 0 && (
          <div className="text-center py-12">
            <Target className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">No budgets found</p>
            <p className="text-sm text-gray-500 mt-2">
              {searchTerm || statusFilter !== 'all' || fiscalYearFilter !== 'all'
                ? 'Try adjusting your filters'
                : 'Create your first budget to get started'
              }
            </p>
            <button
              onClick={openCreateModal}
              className="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-4 h-4" />
              <span>Create Budget</span>
            </button>
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      {(showCreateModal || showEditModal) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between p-6 border-b">
              <h2 className="text-2xl font-bold text-gray-900">
                {showCreateModal ? 'Create New Budget' : 'Edit Budget'}
              </h2>
              <button
                onClick={closeAllModals}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <XCircle className="w-6 h-6" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Budget Name *
                  </label>
                  <input
                    type="text"
                    value={formData.budget_name}
                    onChange={(e) => setFormData({...formData, budget_name: e.target.value})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                    placeholder="e.g., Marketing Budget 2024"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Fiscal Year *
                    </label>
                    <input
                      type="number"
                      value={formData.fiscal_year}
                      onChange={(e) => setFormData({...formData, fiscal_year: parseInt(e.target.value)})}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Period Type *
                    </label>
                    <select
                      value={formData.period_type}
                      onChange={(e) => setFormData({...formData, period_type: e.target.value as any})}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                    >
                      <option value="monthly">Monthly</option>
                      <option value="quarterly">Quarterly</option>
                      <option value="yearly">Yearly</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Account *
                  </label>
                  <select
                    value={formData.account_id}
                    onChange={(e) => setFormData({...formData, account_id: e.target.value})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  >
                    <option value="">Select Account</option>
                    {accounts.map(account => (
                      <option key={account.id} value={account.id}>
                        {account.account_code} - {account.account_name}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Department
                  </label>
                  <input
                    type="text"
                    value={formData.department}
                    onChange={(e) => setFormData({...formData, department: e.target.value})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                    placeholder="e.g., Marketing, Sales, IT"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Total Budget Amount *
                  </label>
                  <input
                    type="number"
                    value={formData.total_amount}
                    onChange={(e) => setFormData({...formData, total_amount: parseFloat(e.target.value)})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                    placeholder="0"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Start Date *
                    </label>
                    <input
                      type="date"
                      value={formData.start_date}
                      onChange={(e) => setFormData({...formData, start_date: e.target.value})}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      End Date *
                    </label>
                    <input
                      type="date"
                      value={formData.end_date}
                      onChange={(e) => setFormData({...formData, end_date: e.target.value})}
                      className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Status
                  </label>
                  <select
                    value={formData.status}
                    onChange={(e) => setFormData({...formData, status: e.target.value})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                  >
                    <option value="draft">Draft</option>
                    <option value="submitted">Submitted</option>
                  </select>
                </div>
              </div>
            </div>

            <div className="flex gap-3 p-6 border-t">
              <button
                onClick={closeAllModals}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={showCreateModal ? handleCreateBudget : handleUpdateBudget}
                className="flex-1 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
              >
                {showCreateModal ? 'Create Budget' : 'Update Budget'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* View Modal */}
      {showViewModal && selectedBudget && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between p-6 border-b">
              <h2 className="text-2xl font-bold text-gray-900">Budget Details</h2>
              <button
                onClick={closeAllModals}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <XCircle className="w-6 h-6" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
              <div className="space-y-6">
                {/* Header Info */}
                <div className="bg-gradient-to-r from-emerald-50 to-blue-50 p-6 rounded-lg">
                  <h3 className="text-2xl font-bold text-gray-900 mb-2">{selectedBudget.budget_name}</h3>
                  <div className="flex items-center gap-4">
                    <span className="text-sm text-gray-600">FY {selectedBudget.fiscal_year}</span>
                    <span className="text-gray-300">•</span>
                    {getPeriodTypeBadge(selectedBudget.period_type)}
                    <span className="text-gray-300">•</span>
                    {getStatusBadge(selectedBudget.status)}
                  </div>
                </div>

                {/* Financial Summary */}
                <div className="grid grid-cols-3 gap-4">
                  <div className="bg-blue-50 p-4 rounded-lg">
                    <p className="text-sm text-blue-700 mb-1">Budget Amount</p>
                    <p className="text-2xl font-bold text-blue-900">{formatCurrency(selectedBudget.total_amount)}</p>
                  </div>
                  <div className="bg-red-50 p-4 rounded-lg">
                    <p className="text-sm text-red-700 mb-1">Spent</p>
                    <p className="text-2xl font-bold text-red-900">{formatCurrency(selectedBudget.spent_amount)}</p>
                    <p className="text-xs text-red-600 mt-1">
                      {calculateUtilization(selectedBudget.spent_amount, selectedBudget.total_amount).toFixed(1)}% used
                    </p>
                  </div>
                  <div className="bg-green-50 p-4 rounded-lg">
                    <p className="text-sm text-green-700 mb-1">Remaining</p>
                    <p className="text-2xl font-bold text-green-900">{formatCurrency(selectedBudget.remaining_amount)}</p>
                    <p className="text-xs text-green-600 mt-1">
                      {(100 - calculateUtilization(selectedBudget.spent_amount, selectedBudget.total_amount)).toFixed(1)}% left
                    </p>
                  </div>
                </div>

                {/* Progress Bar */}
                <div>
                  <div className="flex justify-between text-sm mb-2">
                    <span className="text-gray-600">Budget Utilization</span>
                    <span className="font-semibold text-gray-900">
                      {calculateUtilization(selectedBudget.spent_amount, selectedBudget.total_amount).toFixed(1)}%
                    </span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-4">
                    <div
                      className={`h-4 rounded-full transition-all ${
                        calculateUtilization(selectedBudget.spent_amount, selectedBudget.total_amount) > 100
                          ? 'bg-red-600'
                          : calculateUtilization(selectedBudget.spent_amount, selectedBudget.total_amount) > 80
                          ? 'bg-yellow-600'
                          : 'bg-green-600'
                      }`}
                      style={{
                        width: `${Math.min(calculateUtilization(selectedBudget.spent_amount, selectedBudget.total_amount), 100)}%`
                      }}
                    ></div>
                  </div>
                </div>

                {/* Details Grid */}
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-gray-600 mb-1">Account</p>
                    <p className="font-semibold text-gray-900">{selectedBudget.account_name}</p>
                    <p className="text-xs text-gray-500">{selectedBudget.account_code}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600 mb-1">Department</p>
                    <p className="font-semibold text-gray-900">{selectedBudget.department || '-'}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600 mb-1">Start Date</p>
                    <p className="font-semibold text-gray-900">{formatDate(selectedBudget.start_date)}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600 mb-1">End Date</p>
                    <p className="font-semibold text-gray-900">{formatDate(selectedBudget.end_date)}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600 mb-1">Created By</p>
                    <p className="font-semibold text-gray-900">{selectedBudget.created_by}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600 mb-1">Created At</p>
                    <p className="font-semibold text-gray-900">{formatDate(selectedBudget.created_at)}</p>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex gap-3 p-6 border-t">
              <button
                onClick={closeAllModals}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Close
              </button>
              <button
                onClick={() => handleViewVariance(selectedBudget)}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
              >
                <BarChart3 className="w-4 h-4" />
                <span>View Variance Analysis</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Variance Analysis Modal */}
      {showVarianceModal && selectedBudget && varianceData && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between p-6 border-b">
              <div>
                <h2 className="text-2xl font-bold text-gray-900">Variance Analysis</h2>
                <p className="text-gray-600 text-sm">{selectedBudget.budget_name}</p>
              </div>
              <button
                onClick={closeAllModals}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <XCircle className="w-6 h-6" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto p-6">
              <div className="space-y-6">
                {/* Summary Cards */}
                <div className="grid grid-cols-4 gap-4">
                  <div className="bg-blue-50 p-4 rounded-lg">
                    <p className="text-xs text-blue-700 mb-1">Budgeted</p>
                    <p className="text-xl font-bold text-blue-900">
                      {formatCurrency(varianceData.budgeted_amount || selectedBudget.total_amount)}
                    </p>
                  </div>
                  <div className="bg-red-50 p-4 rounded-lg">
                    <p className="text-xs text-red-700 mb-1">Actual</p>
                    <p className="text-xl font-bold text-red-900">
                      {formatCurrency(varianceData.actual_amount || selectedBudget.spent_amount)}
                    </p>
                  </div>
                  <div className={`p-4 rounded-lg ${
                    (varianceData.variance_amount || 0) > 0 ? 'bg-red-50' : 'bg-green-50'
                  }`}>
                    <p className={`text-xs mb-1 ${
                      (varianceData.variance_amount || 0) > 0 ? 'text-red-700' : 'text-green-700'
                    }`}>
                      Variance
                    </p>
                    <p className={`text-xl font-bold ${
                      (varianceData.variance_amount || 0) > 0 ? 'text-red-900' : 'text-green-900'
                    }`}>
                      {formatCurrency(Math.abs(varianceData.variance_amount || 0))}
                    </p>
                  </div>
                  <div className="bg-purple-50 p-4 rounded-lg">
                    <p className="text-xs text-purple-700 mb-1">Variance %</p>
                    <div className="flex items-center gap-2">
                      <p className="text-xl font-bold text-purple-900">
                        {Math.abs(varianceData.variance_percentage || 0).toFixed(1)}%
                      </p>
                      {getVarianceIndicator(
                        varianceData.variance_amount || 0,
                        varianceData.variance_percentage || 0
                      )}
                    </div>
                  </div>
                </div>

                {/* Variance Chart Placeholder */}
                <div className="bg-gray-50 p-8 rounded-lg text-center">
                  <BarChart3 className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600 font-semibold">Budget vs Actual Comparison</p>
                  <p className="text-sm text-gray-500 mt-2">
                    Visual chart will be displayed here
                  </p>
                  <div className="mt-6 grid grid-cols-2 gap-4 max-w-md mx-auto">
                    <div className="text-left">
                      <div className="flex items-center gap-2 mb-2">
                        <div className="w-4 h-4 bg-blue-500 rounded"></div>
                        <span className="text-sm font-medium">Budgeted</span>
                      </div>
                      <p className="text-2xl font-bold text-blue-600">
                        {formatCurrency(varianceData.budgeted_amount || selectedBudget.total_amount)}
                      </p>
                    </div>
                    <div className="text-left">
                      <div className="flex items-center gap-2 mb-2">
                        <div className="w-4 h-4 bg-red-500 rounded"></div>
                        <span className="text-sm font-medium">Actual</span>
                      </div>
                      <p className="text-2xl font-bold text-red-600">
                        {formatCurrency(varianceData.actual_amount || selectedBudget.spent_amount)}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Analysis */}
                <div className="bg-white border border-gray-200 rounded-lg p-6">
                  <h3 className="font-semibold text-gray-900 mb-4">Variance Analysis</h3>
                  <div className="space-y-3">
                    {(varianceData.variance_amount || 0) > 0 ? (
                      <div className="flex items-start gap-3 p-4 bg-red-50 rounded-lg">
                        <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
                        <div>
                          <p className="font-semibold text-red-900">Over Budget</p>
                          <p className="text-sm text-red-700 mt-1">
                            Spending has exceeded the budget by {formatCurrency(Math.abs(varianceData.variance_amount || 0))} 
                            ({Math.abs(varianceData.variance_percentage || 0).toFixed(1)}%). 
                            Review spending and consider budget adjustment.
                          </p>
                        </div>
                      </div>
                    ) : (varianceData.variance_amount || 0) < 0 ? (
                      <div className="flex items-start gap-3 p-4 bg-green-50 rounded-lg">
                        <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                        <div>
                          <p className="font-semibold text-green-900">Under Budget</p>
                          <p className="text-sm text-green-700 mt-1">
                            Spending is {formatCurrency(Math.abs(varianceData.variance_amount || 0))} 
                            ({Math.abs(varianceData.variance_percentage || 0).toFixed(1)}%) under budget. 
                            Good cost management!
                          </p>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-start gap-3 p-4 bg-blue-50 rounded-lg">
                        <Target className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                        <div>
                          <p className="font-semibold text-blue-900">On Budget</p>
                          <p className="text-sm text-blue-700 mt-1">
                            Spending is exactly on budget. Excellent budget management!
                          </p>
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>

            <div className="flex gap-3 p-6 border-t">
              <button
                onClick={closeAllModals}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Close
              </button>
              <button
                onClick={() => toast.info('Export feature coming soon')}
                className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                <Download className="w-4 h-4" />
                <span>Export Report</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
