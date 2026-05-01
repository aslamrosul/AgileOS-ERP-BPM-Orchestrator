'use client';

import { useEffect, useState } from 'react';
import { 
  ChevronRight,
  ChevronDown,
  Plus,
  Edit,
  Trash2,
  Search,
  Filter,
  Download,
  Upload,
  Eye,
  EyeOff,
  DollarSign,
  TrendingUp,
  TrendingDown,
  Minus,
  RefreshCw,
  FileText,
  AlertCircle
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import AccountForm from '@/components/accounting/AccountForm';

interface Account {
  id: string;
  account_code: string;
  account_name: string;
  account_type: 'asset' | 'liability' | 'equity' | 'revenue' | 'expense';
  parent_account: string | null;
  level: number;
  is_active: boolean;
  currency: string;
  opening_balance: number;
  current_balance: number;
  is_control_account: boolean;
  allow_posting: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
}

interface AccountTreeNode extends Account {
  children: AccountTreeNode[];
  expanded: boolean;
}

export default function ChartOfAccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [accountTree, setAccountTree] = useState<AccountTreeNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [showInactive, setShowInactive] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [modalMode, setModalMode] = useState<'view' | 'create' | 'edit'>('view');

  useEffect(() => {
    fetchAccounts();
  }, []);

  useEffect(() => {
    buildAccountTree();
  }, [accounts, showInactive, typeFilter, searchTerm]);

  const fetchAccounts = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts`
      );
      
      if (!response.ok) {
        throw new Error('Failed to fetch accounts');
      }

      const data = await response.json();
      setAccounts(data || []);
    } catch (error) {
      console.error('Failed to fetch accounts:', error);
      toast.error('Failed to load chart of accounts');
    } finally {
      setLoading(false);
    }
  };

  const buildAccountTree = () => {
    let filtered = [...accounts];

    // Filter by active status
    if (!showInactive) {
      filtered = filtered.filter(acc => acc.is_active);
    }

    // Filter by type
    if (typeFilter !== 'all') {
      filtered = filtered.filter(acc => acc.account_type === typeFilter);
    }

    // Filter by search term
    if (searchTerm) {
      filtered = filtered.filter(acc =>
        acc.account_code.toLowerCase().includes(searchTerm.toLowerCase()) ||
        acc.account_name.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Build tree structure
    const accountMap = new Map<string, AccountTreeNode>();
    const rootAccounts: AccountTreeNode[] = [];

    // First pass: create all nodes
    filtered.forEach(account => {
      accountMap.set(account.id, {
        ...account,
        children: [],
        expanded: account.level === 1 // Expand level 1 by default
      });
    });

    // Second pass: build hierarchy
    filtered.forEach(account => {
      const node = accountMap.get(account.id)!;
      if (account.parent_account && accountMap.has(account.parent_account)) {
        const parent = accountMap.get(account.parent_account)!;
        parent.children.push(node);
      } else {
        rootAccounts.push(node);
      }
    });

    // Sort by account code
    const sortByCode = (a: AccountTreeNode, b: AccountTreeNode) => 
      a.account_code.localeCompare(b.account_code);

    rootAccounts.sort(sortByCode);
    rootAccounts.forEach(node => sortChildren(node));

    setAccountTree(rootAccounts);
  };

  const sortChildren = (node: AccountTreeNode) => {
    if (node.children.length > 0) {
      node.children.sort((a, b) => a.account_code.localeCompare(b.account_code));
      node.children.forEach(child => sortChildren(child));
    }
  };

  const toggleExpand = (accountId: string) => {
    const toggleNode = (nodes: AccountTreeNode[]): AccountTreeNode[] => {
      return nodes.map(node => {
        if (node.id === accountId) {
          return { ...node, expanded: !node.expanded };
        }
        if (node.children.length > 0) {
          return { ...node, children: toggleNode(node.children) };
        }
        return node;
      });
    };

    setAccountTree(toggleNode(accountTree));
  };

  const expandAll = () => {
    const expandNodes = (nodes: AccountTreeNode[]): AccountTreeNode[] => {
      return nodes.map(node => ({
        ...node,
        expanded: true,
        children: expandNodes(node.children)
      }));
    };
    setAccountTree(expandNodes(accountTree));
  };

  const collapseAll = () => {
    const collapseNodes = (nodes: AccountTreeNode[]): AccountTreeNode[] => {
      return nodes.map(node => ({
        ...node,
        expanded: node.level === 1,
        children: collapseNodes(node.children)
      }));
    };
    setAccountTree(collapseNodes(accountTree));
  };

  const handleCreateAccount = () => {
    setSelectedAccount(null);
    setModalMode('create');
    setShowModal(true);
  };

  const handleEditAccount = (account: Account) => {
    setSelectedAccount(account);
    setModalMode('edit');
    setShowModal(true);
  };

  const handleViewAccount = (account: Account) => {
    setSelectedAccount(account);
    setModalMode('view');
    setShowModal(true);
  };

  const handleDeleteAccount = async (accountId: string) => {
    if (!confirm('Are you sure you want to deactivate this account?')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts/${accountId}`,
        { method: 'DELETE' }
      );

      if (!response.ok) {
        throw new Error('Failed to delete account');
      }

      toast.success('Account deactivated successfully');
      fetchAccounts();
    } catch (error) {
      console.error('Failed to delete account:', error);
      toast.error('Failed to deactivate account');
    }
  };

  const getAccountTypeIcon = (type: string) => {
    switch (type) {
      case 'asset':
        return <TrendingUp className="w-4 h-4 text-blue-600" />;
      case 'liability':
        return <TrendingDown className="w-4 h-4 text-red-600" />;
      case 'equity':
        return <DollarSign className="w-4 h-4 text-emerald-600" />;
      case 'revenue':
        return <Plus className="w-4 h-4 text-green-600" />;
      case 'expense':
        return <Minus className="w-4 h-4 text-orange-600" />;
      default:
        return <FileText className="w-4 h-4 text-gray-600" />;
    }
  };

  const getAccountTypeBadge = (type: string) => {
    const badges: Record<string, { bg: string; text: string; label: string }> = {
      asset: { bg: 'bg-blue-100', text: 'text-blue-800', label: 'Asset' },
      liability: { bg: 'bg-red-100', text: 'text-red-800', label: 'Liability' },
      equity: { bg: 'bg-emerald-100', text: 'text-emerald-800', label: 'Equity' },
      revenue: { bg: 'bg-green-100', text: 'text-green-800', label: 'Revenue' },
      expense: { bg: 'bg-orange-100', text: 'text-orange-800', label: 'Expense' }
    };
    const badge = badges[type] || { bg: 'bg-gray-100', text: 'text-gray-800', label: type };
    return (
      <span className={`px-2 py-1 text-xs font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        {badge.label}
      </span>
    );
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0
    }).format(amount);
  };

  const renderAccountNode = (node: AccountTreeNode, depth: number = 0) => {
    const hasChildren = node.children.length > 0;
    const indentClass = `pl-${depth * 6}`;

    return (
      <div key={node.id}>
        {/* Account Row */}
        <div
          className={`
            group flex items-center gap-3 px-4 py-3 hover:bg-gray-50 border-b border-gray-100
            ${!node.is_active ? 'opacity-50' : ''}
            ${node.is_control_account ? 'bg-gray-50 font-semibold' : ''}
          `}
          style={{ paddingLeft: `${depth * 24 + 16}px` }}
        >
          {/* Expand/Collapse Button */}
          <button
            onClick={() => toggleExpand(node.id)}
            className={`flex-shrink-0 w-6 h-6 flex items-center justify-center rounded hover:bg-gray-200 transition-colors ${!hasChildren ? 'invisible' : ''}`}
          >
            {hasChildren && (
              node.expanded ? (
                <ChevronDown className="w-4 h-4 text-gray-600" />
              ) : (
                <ChevronRight className="w-4 h-4 text-gray-600" />
              )
            )}
          </button>

          {/* Account Type Icon */}
          <div className="flex-shrink-0">
            {getAccountTypeIcon(node.account_type)}
          </div>

          {/* Account Code & Name */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-mono text-sm font-medium text-gray-900">
                {node.account_code}
              </span>
              <span className="text-sm text-gray-700">
                {node.account_name}
              </span>
              {node.is_control_account && (
                <span className="text-xs text-gray-500 bg-gray-200 px-2 py-0.5 rounded">
                  Control
                </span>
              )}
              {!node.allow_posting && (
                <span className="text-xs text-orange-600 bg-orange-100 px-2 py-0.5 rounded">
                  No Posting
                </span>
              )}
            </div>
          </div>

          {/* Account Type Badge */}
          <div className="flex-shrink-0">
            {getAccountTypeBadge(node.account_type)}
          </div>

          {/* Balance */}
          <div className="flex-shrink-0 w-40 text-right">
            <p className="text-sm font-medium text-gray-900">
              {formatCurrency(node.current_balance)}
            </p>
            <p className="text-xs text-gray-500">
              Opening: {formatCurrency(node.opening_balance)}
            </p>
          </div>

          {/* Actions */}
          <div className="flex-shrink-0 flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onClick={() => handleViewAccount(node)}
              className="p-1.5 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
              title="View Details"
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={() => handleEditAccount(node)}
              className="p-1.5 text-gray-600 hover:text-emerald-600 hover:bg-emerald-50 rounded transition-colors"
              title="Edit Account"
            >
              <Edit className="w-4 h-4" />
            </button>
            {!node.is_control_account && (
              <button
                onClick={() => handleDeleteAccount(node.id)}
                className="p-1.5 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                title="Deactivate Account"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>

        {/* Children */}
        {hasChildren && node.expanded && (
          <div>
            {node.children.map(child => renderAccountNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading chart of accounts...</p>
        </div>
      </div>
    );
  }

  const stats = {
    total: accounts.length,
    active: accounts.filter(a => a.is_active).length,
    inactive: accounts.filter(a => !a.is_active).length,
    assets: accounts.filter(a => a.account_type === 'asset').length,
    liabilities: accounts.filter(a => a.account_type === 'liability').length,
    equity: accounts.filter(a => a.account_type === 'equity').length,
    revenue: accounts.filter(a => a.account_type === 'revenue').length,
    expenses: accounts.filter(a => a.account_type === 'expense').length
  };

  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Chart of Accounts</h1>
            <p className="text-gray-600 mt-2">Manage your account structure and hierarchy</p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => toast.info('Import feature coming soon')}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Upload className="w-4 h-4" />
              <span>Import</span>
            </button>
            <button
              onClick={() => toast.info('Export feature coming soon')}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span>Export</span>
            </button>
            <button
              onClick={handleCreateAccount}
              className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              <span>New Account</span>
            </button>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Total</p>
          <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Active</p>
          <p className="text-2xl font-bold text-green-600">{stats.active}</p>
        </div>
        <div className="bg-blue-50 rounded-lg shadow-sm p-4 border border-blue-200">
          <p className="text-xs text-blue-700 mb-1">Assets</p>
          <p className="text-2xl font-bold text-blue-900">{stats.assets}</p>
        </div>
        <div className="bg-red-50 rounded-lg shadow-sm p-4 border border-red-200">
          <p className="text-xs text-red-700 mb-1">Liabilities</p>
          <p className="text-2xl font-bold text-red-900">{stats.liabilities}</p>
        </div>
        <div className="bg-emerald-50 rounded-lg shadow-sm p-4 border border-emerald-200">
          <p className="text-xs text-emerald-700 mb-1">Equity</p>
          <p className="text-2xl font-bold text-emerald-900">{stats.equity}</p>
        </div>
        <div className="bg-green-50 rounded-lg shadow-sm p-4 border border-green-200">
          <p className="text-xs text-green-700 mb-1">Revenue</p>
          <p className="text-2xl font-bold text-green-900">{stats.revenue}</p>
        </div>
        <div className="bg-orange-50 rounded-lg shadow-sm p-4 border border-orange-200">
          <p className="text-xs text-orange-700 mb-1">Expenses</p>
          <p className="text-2xl font-bold text-orange-900">{stats.expenses}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Inactive</p>
          <p className="text-2xl font-bold text-gray-400">{stats.inactive}</p>
        </div>
      </div>

      {/* Filters & Controls */}
      <div className="bg-white rounded-lg shadow-sm p-6 border border-gray-200 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {/* Search */}
          <div className="md:col-span-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search by code or name..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>
          </div>

          {/* Type Filter */}
          <div>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Types</option>
              <option value="asset">Assets</option>
              <option value="liability">Liabilities</option>
              <option value="equity">Equity</option>
              <option value="revenue">Revenue</option>
              <option value="expense">Expenses</option>
            </select>
          </div>

          {/* Show Inactive Toggle */}
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowInactive(!showInactive)}
              className={`flex items-center gap-2 px-4 py-2 border rounded-lg transition-colors ${
                showInactive 
                  ? 'bg-emerald-50 border-emerald-300 text-emerald-700' 
                  : 'border-gray-300 text-gray-700 hover:bg-gray-50'
              }`}
            >
              {showInactive ? <Eye className="w-4 h-4" /> : <EyeOff className="w-4 h-4" />}
              <span className="text-sm">Show Inactive</span>
            </button>
          </div>
        </div>

        {/* Tree Controls */}
        <div className="mt-4 flex items-center gap-2">
          <button
            onClick={expandAll}
            className="text-sm text-emerald-600 hover:text-emerald-700 font-medium"
          >
            Expand All
          </button>
          <span className="text-gray-300">|</span>
          <button
            onClick={collapseAll}
            className="text-sm text-emerald-600 hover:text-emerald-700 font-medium"
          >
            Collapse All
          </button>
          <span className="text-gray-300">|</span>
          <button
            onClick={fetchAccounts}
            className="flex items-center gap-1 text-sm text-emerald-600 hover:text-emerald-700 font-medium"
          >
            <RefreshCw className="w-3 h-3" />
            Refresh
          </button>
        </div>
      </div>

      {/* Account Tree */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        {/* Table Header */}
        <div className="bg-gray-50 border-b border-gray-200 px-4 py-3">
          <div className="flex items-center gap-3 font-medium text-xs text-gray-700 uppercase tracking-wider">
            <div className="w-6"></div>
            <div className="w-4"></div>
            <div className="flex-1">Account</div>
            <div className="w-32">Type</div>
            <div className="w-40 text-right">Balance</div>
            <div className="w-24"></div>
          </div>
        </div>

        {/* Tree Content */}
        <div className="max-h-[600px] overflow-y-auto">
          {accountTree.length > 0 ? (
            accountTree.map(node => renderAccountNode(node))
          ) : (
            <div className="text-center py-12">
              <AlertCircle className="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600">No accounts found</p>
              <p className="text-sm text-gray-500 mt-2">
                {searchTerm || typeFilter !== 'all' 
                  ? 'Try adjusting your filters' 
                  : 'Create your first account to get started'
                }
              </p>
            </div>
          )}
        </div>
      </div>

      {/* Account Modal */}
      {showModal && (
        <AccountForm
          account={selectedAccount}
          mode={modalMode}
          onClose={() => setShowModal(false)}
          onSuccess={fetchAccounts}
        />
      )}
    </div>
  );
}
