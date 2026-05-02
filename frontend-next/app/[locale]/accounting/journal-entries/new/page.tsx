'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  ArrowLeft,
  Plus,
  Trash2,
  Save,
  AlertCircle,
  CheckCircle2,
  Calendar,
  FileText,
  DollarSign,
  Hash,
  Info,
  Search
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface Account {
  id: string;
  account_code: string;
  account_name: string;
  account_type: string;
  allow_posting: boolean;
}

interface JournalLine {
  id: string;
  account_id: string;
  account_code: string;
  account_name: string;
  debit: number;
  credit: number;
  description: string;
}

export default function NewJournalEntryPage() {
  const router = useRouter();
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Form data
  const [entryDate, setEntryDate] = useState(new Date().toISOString().split('T')[0]);
  const [entryType, setEntryType] = useState<'manual' | 'auto' | 'opening' | 'closing' | 'adjustment'>('manual');
  const [description, setDescription] = useState('');
  const [reference, setReference] = useState('');
  const [lines, setLines] = useState<JournalLine[]>([
    { id: '1', account_id: '', account_code: '', account_name: '', debit: 0, credit: 0, description: '' },
    { id: '2', account_id: '', account_code: '', account_name: '', debit: 0, credit: 0, description: '' }
  ]);

  // Search & filter
  const [searchTerm, setSearchTerm] = useState('');
  const [showAccountSearch, setShowAccountSearch] = useState<string | null>(null);

  useEffect(() => {
    fetchAccounts();
  }, []);

  const fetchAccounts = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts?is_active=true`
      );
      
      if (!response.ok) throw new Error('Failed to fetch accounts');
      
      const data = await response.json();
      // Filter to show only accounts that allow posting
      setAccounts((data || []).filter((acc: Account) => acc.allow_posting));
    } catch (error) {
      console.error('Failed to fetch accounts:', error);
      toast.error('Failed to load accounts');
    } finally {
      setLoading(false);
    }
  };

  const addLine = () => {
    const newLine: JournalLine = {
      id: Date.now().toString(),
      account_id: '',
      account_code: '',
      account_name: '',
      debit: 0,
      credit: 0,
      description: ''
    };
    setLines([...lines, newLine]);
  };

  const removeLine = (lineId: string) => {
    if (lines.length <= 2) {
      toast.error('Journal entry must have at least 2 lines');
      return;
    }
    setLines(lines.filter(line => line.id !== lineId));
  };

  const updateLine = (lineId: string, field: keyof JournalLine, value: any) => {
    setLines(lines.map(line => {
      if (line.id === lineId) {
        const updated = { ...line, [field]: value };
        
        // Auto-clear opposite field when entering debit/credit
        if (field === 'debit' && value > 0) {
          updated.credit = 0;
        } else if (field === 'credit' && value > 0) {
          updated.debit = 0;
        }
        
        return updated;
      }
      return line;
    }));
  };

  const selectAccount = (lineId: string, account: Account) => {
    updateLine(lineId, 'account_id', account.id);
    updateLine(lineId, 'account_code', account.account_code);
    updateLine(lineId, 'account_name', account.account_name);
    setShowAccountSearch(null);
  };

  const calculateTotals = () => {
    const totalDebit = lines.reduce((sum, line) => sum + (line.debit || 0), 0);
    const totalCredit = lines.reduce((sum, line) => sum + (line.credit || 0), 0);
    const difference = totalDebit - totalCredit;
    const isBalanced = Math.abs(difference) < 0.01; // Allow for floating point errors
    
    return { totalDebit, totalCredit, difference, isBalanced };
  };

  const validateForm = (): boolean => {
    // Check required fields
    if (!entryDate) {
      toast.error('Entry date is required');
      return false;
    }

    if (!description.trim()) {
      toast.error('Description is required');
      return false;
    }

    // Check lines
    if (lines.length < 2) {
      toast.error('Journal entry must have at least 2 lines');
      return false;
    }

    // Check each line
    for (const line of lines) {
      if (!line.account_id) {
        toast.error('All lines must have an account selected');
        return false;
      }

      if (line.debit === 0 && line.credit === 0) {
        toast.error('All lines must have either debit or credit amount');
        return false;
      }

      if (line.debit > 0 && line.credit > 0) {
        toast.error('A line cannot have both debit and credit amounts');
        return false;
      }
    }

    // Check balance
    const { isBalanced } = calculateTotals();
    if (!isBalanced) {
      toast.error('Journal entry is not balanced. Total debit must equal total credit.');
      return false;
    }

    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      setSaving(true);

      const payload = {
        entry_date: new Date(entryDate).toISOString(),
        entry_type: entryType,
        description,
        reference: reference || undefined,
        lines: lines.map((line, index) => ({
          line_number: index + 1,
          account_id: line.account_id,
          debit: line.debit,
          credit: line.credit,
          description: line.description || description
        }))
      };

      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        }
      );

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to create journal entry');
      }

      toast.success('Journal entry created successfully');
      router.push('/accounting/journal-entries');
    } catch (error: any) {
      console.error('Failed to create journal entry:', error);
      toast.error(error.message || 'Failed to create journal entry');
    } finally {
      setSaving(false);
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

  const { totalDebit, totalCredit, difference, isBalanced } = calculateTotals();

  const filteredAccounts = accounts.filter(acc =>
    acc.account_code.toLowerCase().includes(searchTerm.toLowerCase()) ||
    acc.account_name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading accounts...</p>
        </div>
      </div>
    );
  }

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
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">New Journal Entry</h1>
            <p className="text-gray-600 mt-2">Create a new general ledger transaction</p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/accounting/journal-entries"
              className="px-6 py-2.5 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors font-medium"
            >
              Cancel
            </Link>
            <button
              onClick={handleSubmit}
              disabled={saving || !isBalanced}
              className="flex items-center gap-2 px-6 py-2.5 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                  <span>Saving...</span>
                </>
              ) : (
                <>
                  <Save className="w-4 h-4" />
                  <span>Save as Draft</span>
                </>
              )}
            </button>
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Entry Header */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Entry Information</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Entry Date */}
            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <Calendar className="w-4 h-4 text-gray-500" />
                Entry Date *
              </label>
              <input
                type="date"
                value={entryDate}
                onChange={(e) => setEntryDate(e.target.value)}
                required
                className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>

            {/* Entry Type */}
            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <FileText className="w-4 h-4 text-gray-500" />
                Entry Type *
              </label>
              <select
                value={entryType}
                onChange={(e) => setEntryType(e.target.value as any)}
                className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              >
                <option value="manual">Manual</option>
                <option value="auto">Auto</option>
                <option value="opening">Opening</option>
                <option value="closing">Closing</option>
                <option value="adjustment">Adjustment</option>
              </select>
            </div>

            {/* Reference */}
            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <Hash className="w-4 h-4 text-gray-500" />
                Reference
              </label>
              <input
                type="text"
                value={reference}
                onChange={(e) => setReference(e.target.value)}
                placeholder="e.g., INV-2026-001"
                className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>
          </div>

          {/* Description */}
          <div className="mt-6">
            <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
              <FileText className="w-4 h-4 text-gray-500" />
              Description *
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              required
              rows={3}
              placeholder="Enter journal entry description..."
              className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent resize-none"
            />
          </div>
        </div>

        {/* Journal Lines */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="bg-gradient-to-r from-emerald-600 to-teal-600 px-6 py-4">
            <h2 className="text-lg font-semibold text-white">Journal Lines</h2>
            <p className="text-emerald-100 text-sm">Double-entry bookkeeping: Debit = Credit</p>
          </div>

          {/* Table Header */}
          <div className="bg-gray-50 border-b border-gray-200 px-6 py-3">
            <div className="grid grid-cols-12 gap-4 text-xs font-medium text-gray-700 uppercase tracking-wider">
              <div className="col-span-4">Account</div>
              <div className="col-span-2">Description</div>
              <div className="col-span-2 text-right">Debit</div>
              <div className="col-span-2 text-right">Credit</div>
              <div className="col-span-2 text-center">Actions</div>
            </div>
          </div>

          {/* Lines */}
          <div className="divide-y divide-gray-200">
            {lines.map((line, index) => (
              <div key={line.id} className="px-6 py-4 hover:bg-gray-50 transition-colors">
                <div className="grid grid-cols-12 gap-4 items-start">
                  {/* Account Selection */}
                  <div className="col-span-4 relative">
                    <div className="relative">
                      <button
                        type="button"
                        onClick={() => setShowAccountSearch(line.id)}
                        className="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-left hover:border-emerald-500 focus:ring-2 focus:ring-emerald-500 focus:border-transparent transition-colors"
                      >
                        {line.account_code ? (
                          <div>
                            <p className="font-mono text-sm font-medium text-gray-900">{line.account_code}</p>
                            <p className="text-xs text-gray-600 truncate">{line.account_name}</p>
                          </div>
                        ) : (
                          <span className="text-gray-500">Select account...</span>
                        )}
                      </button>

                      {/* Account Search Dropdown */}
                      {showAccountSearch === line.id && (
                        <div className="absolute z-10 mt-2 w-full bg-white border border-gray-300 rounded-lg shadow-lg max-h-64 overflow-hidden">
                          <div className="p-2 border-b border-gray-200">
                            <div className="relative">
                              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                              <input
                                type="text"
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                placeholder="Search accounts..."
                                className="w-full pl-9 pr-3 py-2 border border-gray-300 rounded text-sm focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                                autoFocus
                              />
                            </div>
                          </div>
                          <div className="max-h-48 overflow-y-auto">
                            {filteredAccounts.map((account) => (
                              <button
                                key={account.id}
                                type="button"
                                onClick={() => selectAccount(line.id, account)}
                                className="w-full px-4 py-2 text-left hover:bg-emerald-50 transition-colors"
                              >
                                <p className="font-mono text-sm font-medium text-gray-900">{account.account_code}</p>
                                <p className="text-xs text-gray-600">{account.account_name}</p>
                              </button>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Description */}
                  <div className="col-span-2">
                    <input
                      type="text"
                      value={line.description}
                      onChange={(e) => updateLine(line.id, 'description', e.target.value)}
                      placeholder="Line description..."
                      className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                    />
                  </div>

                  {/* Debit */}
                  <div className="col-span-2">
                    <input
                      type="number"
                      value={line.debit || ''}
                      onChange={(e) => updateLine(line.id, 'debit', parseFloat(e.target.value) || 0)}
                      placeholder="0.00"
                      step="0.01"
                      min="0"
                      className="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-right focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                    />
                  </div>

                  {/* Credit */}
                  <div className="col-span-2">
                    <input
                      type="number"
                      value={line.credit || ''}
                      onChange={(e) => updateLine(line.id, 'credit', parseFloat(e.target.value) || 0)}
                      placeholder="0.00"
                      step="0.01"
                      min="0"
                      className="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-right focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                    />
                  </div>

                  {/* Actions */}
                  <div className="col-span-2 flex items-center justify-center gap-2">
                    <button
                      type="button"
                      onClick={() => removeLine(line.id)}
                      disabled={lines.length <= 2}
                      className="p-2 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      title="Remove Line"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Add Line Button */}
          <div className="px-6 py-4 bg-gray-50 border-t border-gray-200">
            <button
              type="button"
              onClick={addLine}
              className="flex items-center gap-2 px-4 py-2 text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors font-medium"
            >
              <Plus className="w-4 h-4" />
              <span>Add Line</span>
            </button>
          </div>

          {/* Totals */}
          <div className="bg-gray-50 border-t border-gray-200 px-6 py-4">
            <div className="grid grid-cols-12 gap-4">
              <div className="col-span-6"></div>
              <div className="col-span-2 text-right">
                <p className="text-sm font-medium text-gray-700 mb-1">Total Debit</p>
                <p className="text-lg font-bold text-gray-900">{formatCurrency(totalDebit)}</p>
              </div>
              <div className="col-span-2 text-right">
                <p className="text-sm font-medium text-gray-700 mb-1">Total Credit</p>
                <p className="text-lg font-bold text-gray-900">{formatCurrency(totalCredit)}</p>
              </div>
              <div className="col-span-2 text-center">
                {isBalanced ? (
                  <div className="flex items-center justify-center gap-2 text-green-600">
                    <CheckCircle2 className="w-5 h-5" />
                    <span className="text-sm font-medium">Balanced</span>
                  </div>
                ) : (
                  <div className="flex flex-col items-center gap-1">
                    <div className="flex items-center gap-2 text-red-600">
                      <AlertCircle className="w-5 h-5" />
                      <span className="text-sm font-medium">Not Balanced</span>
                    </div>
                    <p className="text-xs text-red-600">
                      Diff: {formatCurrency(Math.abs(difference))}
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Info Box */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <Info className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div className="text-sm text-blue-900">
              <p className="font-medium mb-1">Double-Entry Bookkeeping Rules:</p>
              <ul className="list-disc list-inside space-y-1 text-blue-800">
                <li>Every transaction must have at least two lines (one debit, one credit)</li>
                <li>Total debits must equal total credits</li>
                <li>Each line can have either debit OR credit, not both</li>
                <li>Select accounts that allow direct posting</li>
              </ul>
            </div>
          </div>
        </div>
      </form>

      {/* Click outside to close account search */}
      {showAccountSearch && (
        <div
          className="fixed inset-0 z-0"
          onClick={() => setShowAccountSearch(null)}
        />
      )}
    </div>
  );
}
