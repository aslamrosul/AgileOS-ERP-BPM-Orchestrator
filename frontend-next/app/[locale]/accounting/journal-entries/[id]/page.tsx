'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { 
  ArrowLeft,
  Edit,
  Trash2,
  CheckCircle,
  XCircle,
  Clock,
  FileText,
  Calendar,
  User,
  Hash,
  DollarSign,
  AlertCircle,
  Download,
  Printer,
  Share2,
  RotateCcw,
  Shield,
  Eye
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface JournalLine {
  id: string;
  line_number: number;
  account_id: string;
  account_code: string;
  account_name: string;
  debit: number;
  credit: number;
  description: string;
  cost_center?: string;
  project_id?: string;
  department_id?: string;
}

interface JournalEntry {
  id: string;
  entry_number: string;
  entry_date: string;
  entry_type: 'manual' | 'auto' | 'opening' | 'closing' | 'adjustment';
  description: string;
  reference?: string;
  status: 'draft' | 'posted' | 'reversed';
  lines: JournalLine[];
  total_debit: number;
  total_credit: number;
  posted_by?: string;
  posted_at?: string;
  reversed_by?: string;
  reversed_at?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export default function JournalEntryDetailPage() {
  const params = useParams();
  const router = useRouter();
  const entryId = params.id as string;

  const [entry, setEntry] = useState<JournalEntry | null>(null);
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [showReverseModal, setShowReverseModal] = useState(false);
  const [reverseReason, setReverseReason] = useState('');

  useEffect(() => {
    fetchJournalEntry();
  }, [entryId]);

  const fetchJournalEntry = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries/${entryId}`
      );
      
      if (!response.ok) {
        throw new Error('Failed to fetch journal entry');
      }

      const data = await response.json();
      setEntry(data);
    } catch (error) {
      console.error('Failed to fetch journal entry:', error);
      toast.error('Failed to load journal entry');
      router.push('/accounting/journal-entries');
    } finally {
      setLoading(false);
    }
  };

  const handlePost = async () => {
    if (!confirm('Are you sure you want to post this journal entry? This action cannot be undone.')) {
      return;
    }

    try {
      setProcessing(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries/${entryId}/post`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to post journal entry');
      }

      toast.success('Journal entry posted successfully');
      fetchJournalEntry();
    } catch (error) {
      console.error('Failed to post journal entry:', error);
      toast.error('Failed to post journal entry');
    } finally {
      setProcessing(false);
    }
  };

  const handleReverse = async () => {
    if (!reverseReason.trim()) {
      toast.error('Please provide a reason for reversal');
      return;
    }

    try {
      setProcessing(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries/${entryId}/reverse`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ reason: reverseReason })
        }
      );

      if (!response.ok) {
        throw new Error('Failed to reverse journal entry');
      }

      toast.success('Journal entry reversed successfully');
      setShowReverseModal(false);
      fetchJournalEntry();
    } catch (error) {
      console.error('Failed to reverse journal entry:', error);
      toast.error('Failed to reverse journal entry');
    } finally {
      setProcessing(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this draft journal entry?')) {
      return;
    }

    try {
      setProcessing(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries/${entryId}`,
        { method: 'DELETE' }
      );

      if (!response.ok) {
        throw new Error('Failed to delete journal entry');
      }

      toast.success('Journal entry deleted successfully');
      router.push('/accounting/journal-entries');
    } catch (error) {
      console.error('Failed to delete journal entry:', error);
      toast.error('Failed to delete journal entry');
    } finally {
      setProcessing(false);
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
      <span className={`inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-lg ${badge.bg} ${badge.text}`}>
        <Icon className="w-5 h-5" />
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
      <span className={`px-3 py-1 text-sm font-medium rounded-lg ${badge.bg} ${badge.text}`}>
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
      month: 'long',
      day: 'numeric'
    });
  };

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading journal entry...</p>
        </div>
      </div>
    );
  }

  if (!entry) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <p className="text-gray-600">Journal entry not found</p>
          <Link
            href="/accounting/journal-entries"
            className="mt-4 inline-flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700"
          >
            <ArrowLeft className="w-4 h-4" />
            Back to Journal Entries
          </Link>
        </div>
      </div>
    );
  }

  const isBalanced = Math.abs(entry.total_debit - entry.total_credit) < 0.01;

  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <Link
          href="/accounting/journal-entries"
          className="inline-flex items-center gap-2 text-emerald-600 hover:text-emerald-700 mb-4"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to Journal Entries</span>
        </Link>

        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-4 mb-2">
              <h1 className="text-3xl font-bold text-gray-900">{entry.entry_number}</h1>
              {getStatusBadge(entry.status)}
              {getTypeBadge(entry.entry_type)}
            </div>
            <p className="text-gray-600">{entry.description}</p>
          </div>

          {/* Actions */}
          <div className="flex items-center gap-3">
            <button
              onClick={() => toast.info('Print feature coming soon')}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Printer className="w-4 h-4" />
              <span>Print</span>
            </button>
            <button
              onClick={() => toast.info('Export feature coming soon')}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span>Export</span>
            </button>
            
            {entry.status === 'draft' && (
              <>
                <Link
                  href={`/accounting/journal-entries/${entryId}/edit`}
                  className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <Edit className="w-4 h-4" />
                  <span>Edit</span>
                </Link>
                <button
                  onClick={handlePost}
                  disabled={processing || !isBalanced}
                  className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <CheckCircle className="w-4 h-4" />
                  <span>Post Entry</span>
                </button>
                <button
                  onClick={handleDelete}
                  disabled={processing}
                  className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Trash2 className="w-4 h-4" />
                  <span>Delete</span>
                </button>
              </>
            )}

            {entry.status === 'posted' && (
              <button
                onClick={() => setShowReverseModal(true)}
                disabled={processing}
                className="flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <RotateCcw className="w-4 h-4" />
                <span>Reverse Entry</span>
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Entry Information */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        {/* Main Info */}
        <div className="lg:col-span-2 bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Entry Information</h2>
          <div className="grid grid-cols-2 gap-6">
            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-600 mb-1">
                <Calendar className="w-4 h-4" />
                Entry Date
              </label>
              <p className="text-base font-medium text-gray-900">{formatDate(entry.entry_date)}</p>
            </div>

            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-600 mb-1">
                <Hash className="w-4 h-4" />
                Entry Number
              </label>
              <p className="text-base font-mono font-medium text-gray-900">{entry.entry_number}</p>
            </div>

            {entry.reference && (
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-600 mb-1">
                  <FileText className="w-4 h-4" />
                  Reference
                </label>
                <p className="text-base font-medium text-gray-900">{entry.reference}</p>
              </div>
            )}

            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-600 mb-1">
                <User className="w-4 h-4" />
                Created By
              </label>
              <p className="text-base font-medium text-gray-900">{entry.created_by}</p>
            </div>

            <div className="col-span-2">
              <label className="flex items-center gap-2 text-sm font-medium text-gray-600 mb-1">
                <FileText className="w-4 h-4" />
                Description
              </label>
              <p className="text-base text-gray-900">{entry.description}</p>
            </div>
          </div>
        </div>

        {/* Status & Audit */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Status & Audit</h2>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium text-gray-600 mb-2 block">Status</label>
              {getStatusBadge(entry.status)}
            </div>

            <div>
              <label className="text-sm font-medium text-gray-600 mb-1 block">Created</label>
              <p className="text-sm text-gray-900">{formatDateTime(entry.created_at)}</p>
            </div>

            {entry.posted_at && (
              <div>
                <label className="text-sm font-medium text-gray-600 mb-1 block">Posted</label>
                <p className="text-sm text-gray-900">{formatDateTime(entry.posted_at)}</p>
                <p className="text-xs text-gray-500">by {entry.posted_by}</p>
              </div>
            )}

            {entry.reversed_at && (
              <div>
                <label className="text-sm font-medium text-gray-600 mb-1 block">Reversed</label>
                <p className="text-sm text-gray-900">{formatDateTime(entry.reversed_at)}</p>
                <p className="text-xs text-gray-500">by {entry.reversed_by}</p>
              </div>
            )}

            <div className="pt-4 border-t border-gray-200">
              <div className="flex items-center gap-2 text-sm text-gray-600">
                <Shield className="w-4 h-4" />
                <span>Audit trail enabled</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Journal Lines */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden mb-6">
        <div className="bg-gradient-to-r from-emerald-600 to-teal-600 px-6 py-4">
          <h2 className="text-lg font-semibold text-white">Journal Lines</h2>
          <p className="text-emerald-100 text-sm">Double-entry bookkeeping details</p>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  #
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Account
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Description
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Debit
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Credit
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {entry.lines?.map((line) => (
                <tr key={line.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {line.line_number}
                  </td>
                  <td className="px-6 py-4">
                    <p className="font-mono text-sm font-medium text-gray-900">{line.account_code}</p>
                    <p className="text-sm text-gray-600">{line.account_name}</p>
                  </td>
                  <td className="px-6 py-4">
                    <p className="text-sm text-gray-900">{line.description || '-'}</p>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    {line.debit > 0 ? (
                      <span className="text-sm font-medium text-gray-900">{formatCurrency(line.debit)}</span>
                    ) : (
                      <span className="text-sm text-gray-400">-</span>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    {line.credit > 0 ? (
                      <span className="text-sm font-medium text-gray-900">{formatCurrency(line.credit)}</span>
                    ) : (
                      <span className="text-sm text-gray-400">-</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
            <tfoot className="bg-gray-50 border-t-2 border-gray-300">
              <tr>
                <td colSpan={3} className="px-6 py-4 text-right">
                  <span className="text-sm font-bold text-gray-900">TOTAL</span>
                </td>
                <td className="px-6 py-4 text-right">
                  <span className="text-base font-bold text-gray-900">{formatCurrency(entry.total_debit)}</span>
                </td>
                <td className="px-6 py-4 text-right">
                  <span className="text-base font-bold text-gray-900">{formatCurrency(entry.total_credit)}</span>
                </td>
              </tr>
            </tfoot>
          </table>
        </div>

        {/* Balance Indicator */}
        <div className="bg-gray-50 border-t border-gray-200 px-6 py-4">
          {isBalanced ? (
            <div className="flex items-center justify-center gap-2 text-green-600">
              <CheckCircle className="w-5 h-5" />
              <span className="text-sm font-medium">Entry is balanced (Debit = Credit)</span>
            </div>
          ) : (
            <div className="flex items-center justify-center gap-2 text-red-600">
              <AlertCircle className="w-5 h-5" />
              <span className="text-sm font-medium">
                Entry is not balanced (Difference: {formatCurrency(Math.abs(entry.total_debit - entry.total_credit))})
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Reverse Modal */}
      {showReverseModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-xl shadow-2xl w-full max-w-md">
            <div className="bg-gradient-to-r from-orange-600 to-red-600 px-6 py-4 flex items-center justify-between rounded-t-xl">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
                  <RotateCcw className="w-6 h-6 text-white" />
                </div>
                <div>
                  <h2 className="text-xl font-bold text-white">Reverse Journal Entry</h2>
                  <p className="text-orange-100 text-sm">This action cannot be undone</p>
                </div>
              </div>
            </div>

            <div className="p-6">
              <div className="bg-orange-50 border border-orange-200 rounded-lg p-4 mb-4">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-orange-600 flex-shrink-0 mt-0.5" />
                  <div className="text-sm text-orange-900">
                    <p className="font-medium mb-1">Warning:</p>
                    <p>Reversing this entry will create a new journal entry with opposite debit/credit amounts. The original entry will be marked as reversed.</p>
                  </div>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Reason for Reversal *
                </label>
                <textarea
                  value={reverseReason}
                  onChange={(e) => setReverseReason(e.target.value)}
                  rows={4}
                  placeholder="Enter the reason for reversing this entry..."
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent resize-none"
                />
              </div>

              <div className="flex items-center gap-3 mt-6">
                <button
                  onClick={() => setShowReverseModal(false)}
                  disabled={processing}
                  className="flex-1 px-4 py-2.5 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors font-medium"
                >
                  Cancel
                </button>
                <button
                  onClick={handleReverse}
                  disabled={processing || !reverseReason.trim()}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {processing ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                      <span>Reversing...</span>
                    </>
                  ) : (
                    <>
                      <RotateCcw className="w-4 h-4" />
                      <span>Reverse Entry</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
