'use client';

import { useState, useEffect } from 'react';
import { 
  X, Save, Plus, Trash2, AlertCircle, Info, FileText, Calendar,
  Hash, DollarSign, CheckCircle2, Clock
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';

interface Account {
  id: string;
  account_code: string;
  account_name: string;
  account_type: string;
  allow_posting: boolean;
}

interface JournalEntryLine {
  account_id: string;
  account_code?: string;
  account_name?: string;
  description: string;
  debit: number;
  credit: number;
}

interface JournalEntry {
  id?: string;
  entry_number?: string;
  entry_date: string;
  entry_type: 'manual' | 'auto' | 'opening' | 'closing' | 'adjustment';
  description: string;
  reference?: string;
  status?: 'draft' | 'posted' | 'reversed';
  lines: JournalEntryLine[];
}

interface JournalEntryFormProps {
  entry?: JournalEntry | null;
  mode: 'create' | 'edit' | 'view';
  onClose: () => void;
  onSuccess: () => void;
}

export default function JournalEntryForm({ entry, mode, onClose, onSuccess }: JournalEntryFormProps) {
  const [formData, setFormData] = useState<JournalEntry>({
    entry_date: new Date().toISOString().split('T')[0],
    entry_type: 'manual',
    description: '',
    reference: '',
    lines: [
      { account_id: '', description: '', debit: 0, credit: 0 },
      { account_id: '', description: '', debit: 0, credit: 0 }
    ]
  });

  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<string[]>([]);

  const isReadOnly = mode === 'view' || (entry && entry.status === 'posted');

  useEffect(() => {
    fetchAccounts();
    if (entry) {
      setFormData(entry);
    }
  }, [entry]);

  const fetchAccounts = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts`
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

  const calculateTotals = () => {
    const totalDebit = formData.lines.reduce((sum, line) => sum + (line.debit || 0), 0);
    const totalCredit = formData.lines.reduce((sum, line) => sum + (line.credit || 0), 0);
    const difference = totalDebit - totalCredit;
    return { totalDebit, totalCredit, difference, isBalanced: Math.abs(difference) < 0.01 };
  };

  const validateForm = (): boolean => {
    const newErrors: string[] = [];

    // Basic validation
    if (!formData.entry_date) {
      newErrors.push('Entry date is required');
    }
    if (!formData.description.trim()) {
      newErrors.push('Description is required');
    }

    // Line validation
    if (formData.lines.length < 2) {
      newErrors.push('At least 2 lines are required');
    }

    formData.lines.forEach((line, index) => {
      if (!line.account_id) {
        newErrors.push(`Line ${index + 1}: Account is required`);
      }
      if (!line.description.trim()) {
        newErrors.push(`Line ${index + 1}: Description is required`);
      }
      if (line.debit === 0 && line.credit === 0) {
        newErrors.push(`Line ${index + 1}: Either debit or credit must be greater than 0`);
      }
      if (line.debit > 0 && line.credit > 0) {
        newErrors.push(`Line ${index + 1}: Cannot have both debit and credit`);
      }
    });

    // Balance validation
    const { isBalanced } = calculateTotals();
    if (!isBalanced) {
      newErrors.push('Journal entry must be balanced (Total Debit = Total Credit)');
    }

    setErrors(newErrors);
    return newErrors.length === 0;
  };

  const handleLineChange = (index: number, field: keyof JournalEntryLine, value: any) => {
    const newLines = [...formData.lines];
    newLines[index] = { ...newLines[index], [field]: value };

    // If account changed, update account code and name
    if (field === 'account_id' && value) {
      const account = accounts.find(a => a.id === value);
      if (account) {
        newLines[index].account_code = account.account_code;
        newLines[index].account_name = account.account_name;
      }
    }

    // Auto-clear opposite amount when entering debit/credit
    if (field === 'debit' && value > 0) {
      newLines[index].credit = 0;
    } else if (field === 'credit' && value > 0) {
      newLines[index].debit = 0;
    }

    setFormData({ ...formData, lines: newLines });
  };

  const addLine = () => {
    setFormData({
      ...formData,
      lines: [...formData.lines, { account_id: '', description: '', debit: 0, credit: 0 }]
    });
  };

  const removeLine = (index: number) => {
    if (formData.lines.length <= 2) {
      toast.error('At least 2 lines are required');
      return;
    }
    const newLines = formData.lines.filter((_, i) => i !== index);
    setFormData({ ...formData, lines: newLines });
  };

  const handleSubmit = async (postImmediately: boolean = false) => {
    if (!validateForm()) {
      toast.error('Please fix validation errors');
      return;
    }

    try {
      setSaving(true);

      const url = mode === 'edit' && entry?.id
        ? `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries/${entry.id}`
        : `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries`;

      const method = mode === 'edit' ? 'PUT' : 'POST';

      const response = await authenticatedFetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to save journal entry');
      }

      const result = await response.json();

      // If post immediately, post the entry
      if (postImmediately && result.id) {
        const postResponse = await authenticatedFetch(
          `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/journal-entries/${result.id}/post`,
          { method: 'POST' }
        );

        if (!postResponse.ok) {
          throw new Error('Entry saved but failed to post');
        }
      }

      toast.success(
        postImmediately 
          ? 'Journal entry posted successfully' 
          : `Journal entry ${mode === 'edit' ? 'updated' : 'created'} successfully`
      );
      onSuccess();
      onClose();
    } catch (error: any) {
      console.error('Failed to save journal entry:', error);
      toast.error(error.message || 'Failed to save journal entry');
    } finally {
      setSaving(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(amount);
  };

  const { totalDebit, totalCredit, difference, isBalanced } = calculateTotals();

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl shadow-2xl w-full max-w-6xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="bg-gradient-to-r from-emerald-600 to-teal-600 px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
              <FileText className="w-6 h-6 text-white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white">
                {mode === 'create' && 'Create Journal Entry'}
                {mode === 'edit' && 'Edit Journal Entry'}
                {mode === 'view' && 'Journal Entry Details'}
              </h2>
              <p className="text-emerald-100 text-sm">
                {mode === 'create' && 'Record a new journal entry'}
                {mode === 'edit' && 'Update journal entry information'}
                {mode === 'view' && 'View journal entry details'}
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-white hover:bg-white hover:bg-opacity-20 rounded-lg p-2 transition-colors"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Form Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="space-y-6">
            {/* Entry Header */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Entry Date */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <Calendar className="w-4 h-4 text-gray-500" />
                  Entry Date *
                </label>
                <input
                  type="date"
                  value={formData.entry_date}
                  onChange={(e) => setFormData({ ...formData, entry_date: e.target.value })}
                  disabled={isReadOnly}
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-500"
                />
              </div>

              {/* Entry Type */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <Hash className="w-4 h-4 text-gray-500" />
                  Entry Type *
                </label>
                <select
                  value={formData.entry_type}
                  onChange={(e) => setFormData({ ...formData, entry_type: e.target.value as any })}
                  disabled={isReadOnly}
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-500"
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
                  value={formData.reference || ''}
                  onChange={(e) => setFormData({ ...formData, reference: e.target.value })}
                  disabled={isReadOnly}
                  placeholder="e.g., INV-2024-001"
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-500"
                />
              </div>
            </div>

            {/* Description */}
            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                <FileText className="w-4 h-4 text-gray-500" />
                Description *
              </label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                disabled={isReadOnly}
                placeholder="Enter journal entry description..."
                rows={2}
                className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-500"
              />
            </div>

            {/* Journal Lines */}
            <div>
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900">Journal Lines</h3>
                {!isReadOnly && (
                  <button
                    onClick={addLine}
                    className="flex items-center gap-2 px-4 py-2 text-sm bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
                  >
                    <Plus className="w-4 h-4" />
                    Add Line
                  </button>
                )}
              </div>

              <div className="bg-gray-50 rounded-lg p-4 overflow-x-auto">
                <table className="w-full min-w-[800px]">
                  <thead>
                    <tr className="border-b border-gray-300">
                      <th className="text-left text-xs font-medium text-gray-700 uppercase tracking-wider pb-2 px-2">
                        Account
                      </th>
                      <th className="text-left text-xs font-medium text-gray-700 uppercase tracking-wider pb-2 px-2">
                        Description
                      </th>
                      <th className="text-right text-xs font-medium text-gray-700 uppercase tracking-wider pb-2 px-2">
                        Debit
                      </th>
                      <th className="text-right text-xs font-medium text-gray-700 uppercase tracking-wider pb-2 px-2">
                        Credit
                      </th>
                      {!isReadOnly && (
                        <th className="text-center text-xs font-medium text-gray-700 uppercase tracking-wider pb-2 px-2 w-16">
                          Action
                        </th>
                      )}
                    </tr>
                  </thead>
                  <tbody>
                    {formData.lines.map((line, index) => (
                      <tr key={index} className="border-b border-gray-200">
                        <td className="py-2 px-2">
                          <select
                            value={line.account_id}
                            onChange={(e) => handleLineChange(index, 'account_id', e.target.value)}
                            disabled={isReadOnly}
                            className="w-full px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-emerald-500 disabled:bg-gray-100"
                          >
                            <option value="">Select Account</option>
                            {accounts.map((account) => (
                              <option key={account.id} value={account.id}>
                                {account.account_code} - {account.account_name}
                              </option>
                            ))}
                          </select>
                        </td>
                        <td className="py-2 px-2">
                          <input
                            type="text"
                            value={line.description}
                            onChange={(e) => handleLineChange(index, 'description', e.target.value)}
                            disabled={isReadOnly}
                            placeholder="Line description"
                            className="w-full px-3 py-2 text-sm border border-gray-300 rounded focus:ring-2 focus:ring-emerald-500 disabled:bg-gray-100"
                          />
                        </td>
                        <td className="py-2 px-2">
                          <input
                            type="number"
                            value={line.debit || ''}
                            onChange={(e) => handleLineChange(index, 'debit', parseFloat(e.target.value) || 0)}
                            disabled={isReadOnly}
                            step="0.01"
                            min="0"
                            placeholder="0.00"
                            className="w-full px-3 py-2 text-sm text-right border border-gray-300 rounded focus:ring-2 focus:ring-emerald-500 disabled:bg-gray-100"
                          />
                        </td>
                        <td className="py-2 px-2">
                          <input
                            type="number"
                            value={line.credit || ''}
                            onChange={(e) => handleLineChange(index, 'credit', parseFloat(e.target.value) || 0)}
                            disabled={isReadOnly}
                            step="0.01"
                            min="0"
                            placeholder="0.00"
                            className="w-full px-3 py-2 text-sm text-right border border-gray-300 rounded focus:ring-2 focus:ring-emerald-500 disabled:bg-gray-100"
                          />
                        </td>
                        {!isReadOnly && (
                          <td className="py-2 px-2 text-center">
                            <button
                              onClick={() => removeLine(index)}
                              disabled={formData.lines.length <= 2}
                              className="p-1.5 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                              title="Remove Line"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                  <tfoot>
                    <tr className="border-t-2 border-gray-400 font-semibold">
                      <td colSpan={2} className="py-3 px-2 text-right text-gray-900">
                        Total:
                      </td>
                      <td className="py-3 px-2 text-right text-gray-900">
                        {formatCurrency(totalDebit)}
                      </td>
                      <td className="py-3 px-2 text-right text-gray-900">
                        {formatCurrency(totalCredit)}
                      </td>
                      {!isReadOnly && <td></td>}
                    </tr>
                    {!isBalanced && (
                      <tr>
                        <td colSpan={2} className="py-2 px-2 text-right text-red-600 font-medium">
                          Difference:
                        </td>
                        <td colSpan={2} className="py-2 px-2 text-right text-red-600 font-medium">
                          {formatCurrency(Math.abs(difference))} {difference > 0 ? '(Debit)' : '(Credit)'}
                        </td>
                        {!isReadOnly && <td></td>}
                      </tr>
                    )}
                  </tfoot>
                </table>
              </div>

              {/* Balance Status */}
              <div className={`mt-4 p-4 rounded-lg border-2 ${isBalanced ? 'bg-green-50 border-green-300' : 'bg-red-50 border-red-300'}`}>
                <div className="flex items-center gap-3">
                  {isBalanced ? (
                    <>
                      <CheckCircle2 className="w-6 h-6 text-green-600 flex-shrink-0" />
                      <div>
                        <p className="font-semibold text-green-900">Entry is Balanced</p>
                        <p className="text-sm text-green-700">Total Debit equals Total Credit</p>
                      </div>
                    </>
                  ) : (
                    <>
                      <AlertCircle className="w-6 h-6 text-red-600 flex-shrink-0" />
                      <div>
                        <p className="font-semibold text-red-900">Entry is Not Balanced</p>
                        <p className="text-sm text-red-700">
                          Difference: {formatCurrency(Math.abs(difference))} {difference > 0 ? '(Debit exceeds Credit)' : '(Credit exceeds Debit)'}
                        </p>
                      </div>
                    </>
                  )}
                </div>
              </div>
            </div>

            {/* Validation Errors */}
            {errors.length > 0 && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
                  <div className="flex-1">
                    <p className="font-semibold text-red-900 mb-2">Validation Errors:</p>
                    <ul className="list-disc list-inside space-y-1 text-sm text-red-700">
                      {errors.map((error, index) => (
                        <li key={index}>{error}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            )}

            {/* Info Box */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <Info className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                <div className="text-sm text-blue-900">
                  <p className="font-medium mb-1">Journal Entry Guidelines:</p>
                  <ul className="list-disc list-inside space-y-1 text-blue-800">
                    <li>Every journal entry must have at least 2 lines</li>
                    <li>Total Debit must equal Total Credit (balanced entry)</li>
                    <li>Each line must have either a debit or credit amount (not both)</li>
                    <li>Draft entries can be edited, posted entries cannot be modified</li>
                    <li>Posted entries can only be reversed, not deleted</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="bg-gray-50 px-6 py-4 flex items-center justify-between border-t border-gray-200">
          <div className="text-sm text-gray-600">
            {errors.length > 0 && (
              <span className="text-red-600 flex items-center gap-1">
                <AlertCircle className="w-4 h-4" />
                {errors.length} validation error{errors.length > 1 ? 's' : ''}
              </span>
            )}
          </div>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={onClose}
              className="px-6 py-2.5 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors font-medium"
            >
              {isReadOnly ? 'Close' : 'Cancel'}
            </button>
            {!isReadOnly && (
              <>
                <button
                  onClick={() => handleSubmit(false)}
                  disabled={saving}
                  className="flex items-center gap-2 px-6 py-2.5 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {saving ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                      <span>Saving...</span>
                    </>
                  ) : (
                    <>
                      <Clock className="w-4 h-4" />
                      <span>Save as Draft</span>
                    </>
                  )}
                </button>
                <button
                  onClick={() => handleSubmit(true)}
                  disabled={saving || !isBalanced}
                  className="flex items-center gap-2 px-6 py-2.5 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {saving ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                      <span>Posting...</span>
                    </>
                  ) : (
                    <>
                      <CheckCircle2 className="w-4 h-4" />
                      <span>Post Entry</span>
                    </>
                  )}
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
