'use client';

import { useEffect, useState } from 'react';
import { 
  Plus,
  Search,
  Filter,
  Download,
  Eye,
  Edit,
  CheckCircle,
  XCircle,
  Clock,
  FileText,
  Calendar,
  User,
  DollarSign,
  AlertCircle,
  RefreshCw,
  ArrowUpRight
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface JournalEntry {
  id: string;
  entry_number: string;
  entry_date: string;
  entry_type: 'manual' | 'auto' | 'opening' | 'closing' | 'adjustment';
  description: string;
  reference?: string;
  status: 'draft' | 'posted' | 'reversed';
  total_debit: number;
  total_credit: number;
  posted_by?: string;
  posted_at?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export default function JournalEntriesPage() {
  const [entries, setEntries] = useState<JournalEntry[]>([]);
  const [filteredEntries, setFilteredEntries] = useState<JournalEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  useEffect(() => {
    fetchJournalEntries();
  }, []);

  useEffect(() => {
    filterEntries();
  }, [entries, searchTerm, statusFilter, typeFilter, dateFrom, dateTo]);

  const fetchJournalEntries = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries`
      );
      
      if (!response.ok) {
        throw new Error('Failed to fetch journal entries');
      }

      const data = await response.json();
      setEntries(data || []);
    } catch (error) {
      console.error('Failed to fetch journal entries:', error);
      toast.error('Failed to load journal entries');
    } finally {
      setLoading(false);
    }
  };

  const filterEntries = () => {
    let filtered = [...entries];

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(entry =>
        entry.entry_number.toLowerCase().includes(searchTerm.toLowerCase()) ||
        entry.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (entry.reference && entry.reference.toLowerCase().includes(searchTerm.toLowerCase()))
      );
    }

    // Status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(entry => entry.status === statusFilter);
    }

    // Type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter(entry => entry.entry_type === typeFilter);
    }

    // Date range filter
    if (dateFrom) {
      filtered = filtered.filter(entry => 
        new Date(entry.entry_date) >= new Date(dateFrom)
      );
    }
    if (dateTo) {
      filtered = filtered.filter(entry => 
        new Date(entry.entry_date) <= new Date(dateTo)
      );
    }

    setFilteredEntries(filtered);
  };

  const handlePostEntry = async (entryId: string) => {
    if (!confirm('Are you sure you want to post this journal entry? This action cannot be undone.')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries/${entryId}/post`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to post journal entry');
      }

      toast.success('Journal entry posted successfully');
      fetchJournalEntries();
    } catch (error) {
      console.error('Failed to post journal entry:', error);
      toast.error('Failed to post journal entry');
    }
  };

  const getStatusBadge = (status: string) => {
    const badges: Record<string, { bg: string; text: string; icon: any }> = {
      draft: { bg: 'bg-yellow-100', text: 'text-yellow-800', icon: Clock },
      posted: { bg: 'bg-green-100', text: 'text-green-800', icon: CheckCircle },
      reversed: { bg: 'bg-red-100', text: 'text-red-800', icon: XCircle }
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

  const getTypeBadge = (type: string) => {
    const badges: Record<string, { bg: string; text: string }> = {
      manual: { bg: 'bg-blue-100', text: 'text-blue-800' },
      auto: { bg: 'bg-purple-100', text: 'text-purple-800' },
      opening: { bg: 'bg-emerald-100', text: 'text-emerald-800' },
      closing: { bg: 'bg-orange-100', text: 'text-orange-800' },
      adjustment: { bg: 'bg-pink-100', text: 'text-pink-800' }
    };
    const badge = badges[type] || { bg: 'bg-gray-100', text: 'text-gray-800' };
    
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded ${badge.bg} ${badge.text}`}>
        {type.charAt(0).toUpperCase() + type.slice(1)}
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading journal entries...</p>
        </div>
      </div>
    );
  }

  const stats = {
    total: entries.length,
    draft: entries.filter(e => e.status === 'draft').length,
    posted: entries.filter(e => e.status === 'posted').length,
    reversed: entries.filter(e => e.status === 'reversed').length,
    totalDebit: entries.reduce((sum, e) => sum + e.total_debit, 0),
    totalCredit: entries.reduce((sum, e) => sum + e.total_credit, 0)
  };

  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Journal Entries</h1>
            <p className="text-gray-600 mt-2">Manage general ledger transactions</p>
          </div>
          <div className="flex items-center gap-3">
            <button
              onClick={() => toast.info('Export feature coming soon')}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span>Export</span>
            </button>
            <Link
              href="/accounting/journal-entries/new"
              className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              <span>New Entry</span>
            </Link>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-6 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Total Entries</p>
          <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
        </div>
        <div className="bg-yellow-50 rounded-lg shadow-sm p-4 border border-yellow-200">
          <p className="text-xs text-yellow-700 mb-1">Draft</p>
          <p className="text-2xl font-bold text-yellow-900">{stats.draft}</p>
        </div>
        <div className="bg-green-50 rounded-lg shadow-sm p-4 border border-green-200">
          <p className="text-xs text-green-700 mb-1">Posted</p>
          <p className="text-2xl font-bold text-green-900">{stats.posted}</p>
        </div>
        <div className="bg-red-50 rounded-lg shadow-sm p-4 border border-red-200">
          <p className="text-xs text-red-700 mb-1">Reversed</p>
          <p className="text-2xl font-bold text-red-900">{stats.reversed}</p>
        </div>
        <div className="bg-blue-50 rounded-lg shadow-sm p-4 border border-blue-200">
          <p className="text-xs text-blue-700 mb-1">Total Debit</p>
          <p className="text-lg font-bold text-blue-900">{formatCurrency(stats.totalDebit)}</p>
        </div>
        <div className="bg-purple-50 rounded-lg shadow-sm p-4 border border-purple-200">
          <p className="text-xs text-purple-700 mb-1">Total Credit</p>
          <p className="text-lg font-bold text-purple-900">{formatCurrency(stats.totalCredit)}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm p-6 border border-gray-200 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          {/* Search */}
          <div className="md:col-span-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search by number, description, or reference..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>
          </div>

          {/* Status Filter */}
          <div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Status</option>
              <option value="draft">Draft</option>
              <option value="posted">Posted</option>
              <option value="reversed">Reversed</option>
            </select>
          </div>

          {/* Type Filter */}
          <div>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Types</option>
              <option value="manual">Manual</option>
              <option value="auto">Auto</option>
              <option value="opening">Opening</option>
              <option value="closing">Closing</option>
              <option value="adjustment">Adjustment</option>
            </select>
          </div>

          {/* Refresh */}
          <div>
            <button
              onClick={fetchJournalEntries}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <RefreshCw className="w-4 h-4" />
              <span>Refresh</span>
            </button>
          </div>
        </div>

        {/* Date Range */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">From Date</label>
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">To Date</label>
            <input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            />
          </div>
        </div>
      </div>

      {/* Journal Entries Table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Entry Number
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Date
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Description
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Amount
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredEntries.map((entry) => (
                <tr key={entry.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <FileText className="w-4 h-4 text-gray-400" />
                      <span className="font-mono text-sm font-medium text-gray-900">
                        {entry.entry_number}
                      </span>
                    </div>
                    {entry.reference && (
                      <p className="text-xs text-gray-500 mt-1">Ref: {entry.reference}</p>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2 text-sm text-gray-900">
                      <Calendar className="w-4 h-4 text-gray-400" />
                      {formatDate(entry.entry_date)}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-900 line-clamp-2">{entry.description}</p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getTypeBadge(entry.entry_type)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {getStatusBadge(entry.status)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="text-sm">
                      <p className="font-medium text-gray-900">{formatCurrency(entry.total_debit)}</p>
                      <p className="text-xs text-gray-500">Debit = Credit</p>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <Link
                        href={`/accounting/journal-entries/${entry.id}`}
                        className="p-1.5 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                        title="View Details"
                      >
                        <Eye className="w-4 h-4" />
                      </Link>
                      {entry.status === 'draft' && (
                        <>
                          <Link
                            href={`/accounting/journal-entries/${entry.id}/edit`}
                            className="p-1.5 text-gray-600 hover:text-emerald-600 hover:bg-emerald-50 rounded transition-colors"
                            title="Edit Entry"
                          >
                            <Edit className="w-4 h-4" />
                          </Link>
                          <button
                            onClick={() => handlePostEntry(entry.id)}
                            className="p-1.5 text-gray-600 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
                            title="Post Entry"
                          >
                            <CheckCircle className="w-4 h-4" />
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredEntries.length === 0 && (
          <div className="text-center py-12">
            <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">No journal entries found</p>
            <p className="text-sm text-gray-500 mt-2">
              {searchTerm || statusFilter !== 'all' || typeFilter !== 'all'
                ? 'Try adjusting your filters'
                : 'Create your first journal entry to get started'
              }
            </p>
            <Link
              href="/accounting/journal-entries/new"
              className="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-4 h-4" />
              <span>Create Journal Entry</span>
            </Link>
          </div>
        )}
      </div>

      {/* Balance Validation Info */}
      {stats.totalDebit !== stats.totalCredit && (
        <div className="mt-6 bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-red-900">Balance Mismatch Detected</p>
              <p className="text-sm text-red-700 mt-1">
                Total Debit ({formatCurrency(stats.totalDebit)}) does not equal Total Credit ({formatCurrency(stats.totalCredit)}).
                Please review your journal entries.
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
