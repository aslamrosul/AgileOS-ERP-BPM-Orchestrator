'use client';

import { useState, useEffect } from 'react';
import { 
  X, 
  Save, 
  AlertCircle, 
  Info,
  DollarSign,
  Hash,
  FileText,
  Layers,
  Lock,
  Unlock,
  CheckCircle2
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';

interface Account {
  id?: string;
  account_code: string;
  account_name: string;
  account_type: 'asset' | 'liability' | 'equity' | 'revenue' | 'expense';
  parent_account: string | null;
  level: number;
  is_active: boolean;
  currency: string;
  opening_balance: number;
  current_balance?: number;
  is_control_account: boolean;
  allow_posting: boolean;
  created_at?: string;
  updated_at?: string;
  created_by?: string;
}

interface AccountFormProps {
  account?: Account | null;
  mode: 'create' | 'edit' | 'view';
  onClose: () => void;
  onSuccess: () => void;
}

interface ValidationError {
  field: string;
  message: string;
}

export default function AccountForm({ account, mode, onClose, onSuccess }: AccountFormProps) {
  const [formData, setFormData] = useState<Account>({
    account_code: '',
    account_name: '',
    account_type: 'asset',
    parent_account: null,
    level: 1,
    is_active: true,
    currency: 'IDR',
    opening_balance: 0,
    is_control_account: false,
    allow_posting: true
  });

  const [parentAccounts, setParentAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<ValidationError[]>([]);
  const [touched, setTouched] = useState<Set<string>>(new Set());

  const isReadOnly = mode === 'view';

  useEffect(() => {
    if (account) {
      setFormData(account);
    }
    fetchParentAccounts();
  }, [account]);

  const fetchParentAccounts = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts?is_active=true`
      );
      
      if (!response.ok) throw new Error('Failed to fetch accounts');
      
      const data = await response.json();
      // Filter to show only control accounts as potential parents
      setParentAccounts((data || []).filter((acc: Account) => acc.is_control_account));
    } catch (error) {
      console.error('Failed to fetch parent accounts:', error);
      toast.error('Failed to load parent accounts');
    } finally {
      setLoading(false);
    }
  };

  const validateForm = (): boolean => {
    const newErrors: ValidationError[] = [];

    // Account Code validation
    if (!formData.account_code.trim()) {
      newErrors.push({ field: 'account_code', message: 'Account code is required' });
    } else if (!/^[0-9]+-[0-9]+$/.test(formData.account_code)) {
      newErrors.push({ field: 'account_code', message: 'Account code must be in format: X-XXXX (e.g., 1-1000)' });
    }

    // Account Name validation
    if (!formData.account_name.trim()) {
      newErrors.push({ field: 'account_name', message: 'Account name is required' });
    } else if (formData.account_name.length < 3) {
      newErrors.push({ field: 'account_name', message: 'Account name must be at least 3 characters' });
    }

    // Level validation
    if (formData.parent_account) {
      const parent = parentAccounts.find(p => p.id === formData.parent_account);
      if (parent && formData.level <= parent.level) {
        newErrors.push({ field: 'level', message: 'Level must be greater than parent level' });
      }
    }

    // Control account validation
    if (formData.is_control_account && formData.allow_posting) {
      newErrors.push({ 
        field: 'allow_posting', 
        message: 'Control accounts cannot allow direct posting' 
      });
    }

    setErrors(newErrors);
    return newErrors.length === 0;
  };

  const handleChange = (field: keyof Account, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    setTouched(prev => new Set(prev).add(field));

    // Auto-adjust posting permission for control accounts
    if (field === 'is_control_account' && value === true) {
      setFormData(prev => ({ ...prev, allow_posting: false }));
    }

    // Auto-calculate level based on parent
    if (field === 'parent_account' && value) {
      const parent = parentAccounts.find(p => p.id === value);
      if (parent) {
        setFormData(prev => ({ ...prev, level: parent.level + 1 }));
      }
    } else if (field === 'parent_account' && !value) {
      setFormData(prev => ({ ...prev, level: 1 }));
    }
  };

  const handleBlur = (field: string) => {
    setTouched(prev => new Set(prev).add(field));
  };

  const getFieldError = (field: string): string | undefined => {
    if (!touched.has(field)) return undefined;
    return errors.find(e => e.field === field)?.message;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Mark all fields as touched
    setTouched(new Set(Object.keys(formData)));

    if (!validateForm()) {
      toast.error('Please fix validation errors');
      return;
    }

    try {
      setSaving(true);

      const url = mode === 'edit' && account?.id
        ? `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts/${account.id}`
        : `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts`;

      const method = mode === 'edit' ? 'PUT' : 'POST';

      const response = await authenticatedFetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to save account');
      }

      toast.success(`Account ${mode === 'edit' ? 'updated' : 'created'} successfully`);
      onSuccess();
      onClose();
    } catch (error: any) {
      console.error('Failed to save account:', error);
      toast.error(error.message || 'Failed to save account');
    } finally {
      setSaving(false);
    }
  };

  const accountTypeOptions = [
    { value: 'asset', label: 'Asset', icon: '📈', color: 'text-blue-600', description: 'Resources owned by the company' },
    { value: 'liability', label: 'Liability', icon: '📉', color: 'text-red-600', description: 'Obligations owed to others' },
    { value: 'equity', label: 'Equity', icon: '💰', color: 'text-emerald-600', description: 'Owner\'s stake in the company' },
    { value: 'revenue', label: 'Revenue', icon: '💵', color: 'text-green-600', description: 'Income from operations' },
    { value: 'expense', label: 'Expense', icon: '💸', color: 'text-orange-600', description: 'Costs of operations' }
  ];

  const currencyOptions = [
    { value: 'IDR', label: 'IDR - Indonesian Rupiah', symbol: 'Rp' },
    { value: 'USD', label: 'USD - US Dollar', symbol: '$' },
    { value: 'EUR', label: 'EUR - Euro', symbol: '€' },
    { value: 'SGD', label: 'SGD - Singapore Dollar', symbol: 'S$' },
    { value: 'MYR', label: 'MYR - Malaysian Ringgit', symbol: 'RM' }
  ];

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="bg-gradient-to-r from-emerald-600 to-teal-600 px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-white bg-opacity-20 rounded-lg flex items-center justify-center">
              <FileText className="w-6 h-6 text-white" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white">
                {mode === 'create' && 'Create New Account'}
                {mode === 'edit' && 'Edit Account'}
                {mode === 'view' && 'Account Details'}
              </h2>
              <p className="text-emerald-100 text-sm">
                {mode === 'create' && 'Add a new account to your chart of accounts'}
                {mode === 'edit' && 'Update account information'}
                {mode === 'view' && 'View account information'}
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
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6">
          <div className="space-y-6">
            {/* Account Code & Name */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Account Code */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <Hash className="w-4 h-4 text-gray-500" />
                  Account Code *
                </label>
                <input
                  type="text"
                  value={formData.account_code}
                  onChange={(e) => handleChange('account_code', e.target.value)}
                  onBlur={() => handleBlur('account_code')}
                  disabled={isReadOnly || mode === 'edit'}
                  placeholder="e.g., 1-1000"
                  className={`
                    w-full px-4 py-2.5 border rounded-lg font-mono text-sm
                    focus:ring-2 focus:ring-emerald-500 focus:border-transparent
                    disabled:bg-gray-50 disabled:text-gray-500
                    ${getFieldError('account_code') ? 'border-red-500' : 'border-gray-300'}
                  `}
                />
                {getFieldError('account_code') && (
                  <p className="mt-1 text-sm text-red-600 flex items-center gap-1">
                    <AlertCircle className="w-4 h-4" />
                    {getFieldError('account_code')}
                  </p>
                )}
                <p className="mt-1 text-xs text-gray-500">
                  Format: Type-Number (e.g., 1-1000 for assets)
                </p>
              </div>

              {/* Account Name */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <FileText className="w-4 h-4 text-gray-500" />
                  Account Name *
                </label>
                <input
                  type="text"
                  value={formData.account_name}
                  onChange={(e) => handleChange('account_name', e.target.value)}
                  onBlur={() => handleBlur('account_name')}
                  disabled={isReadOnly}
                  placeholder="e.g., Cash in Bank"
                  className={`
                    w-full px-4 py-2.5 border rounded-lg
                    focus:ring-2 focus:ring-emerald-500 focus:border-transparent
                    disabled:bg-gray-50 disabled:text-gray-500
                    ${getFieldError('account_name') ? 'border-red-500' : 'border-gray-300'}
                  `}
                />
                {getFieldError('account_name') && (
                  <p className="mt-1 text-sm text-red-600 flex items-center gap-1">
                    <AlertCircle className="w-4 h-4" />
                    {getFieldError('account_name')}
                  </p>
                )}
              </div>
            </div>

            {/* Account Type */}
            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-3">
                <Layers className="w-4 h-4 text-gray-500" />
                Account Type *
              </label>
              <div className="grid grid-cols-1 md:grid-cols-5 gap-3">
                {accountTypeOptions.map((option) => (
                  <button
                    key={option.value}
                    type="button"
                    onClick={() => !isReadOnly && handleChange('account_type', option.value)}
                    disabled={isReadOnly}
                    className={`
                      p-4 border-2 rounded-lg transition-all
                      ${formData.account_type === option.value
                        ? 'border-emerald-500 bg-emerald-50 shadow-md'
                        : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'
                      }
                      ${isReadOnly ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}
                    `}
                  >
                    <div className="text-2xl mb-2">{option.icon}</div>
                    <div className={`text-sm font-semibold ${option.color}`}>
                      {option.label}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                      {option.description}
                    </div>
                  </button>
                ))}
              </div>
            </div>

            {/* Parent Account & Level */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Parent Account */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <Layers className="w-4 h-4 text-gray-500" />
                  Parent Account
                </label>
                <select
                  value={formData.parent_account || ''}
                  onChange={(e) => handleChange('parent_account', e.target.value || null)}
                  disabled={isReadOnly || loading}
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-500"
                >
                  <option value="">None (Root Level)</option>
                  {parentAccounts
                    .filter(acc => acc.account_type === formData.account_type)
                    .map((acc) => (
                      <option key={acc.id} value={acc.id}>
                        {acc.account_code} - {acc.account_name}
                      </option>
                    ))}
                </select>
                <p className="mt-1 text-xs text-gray-500">
                  Select a parent account to create hierarchy
                </p>
              </div>

              {/* Level */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <Layers className="w-4 h-4 text-gray-500" />
                  Level
                </label>
                <input
                  type="number"
                  value={formData.level}
                  onChange={(e) => handleChange('level', parseInt(e.target.value))}
                  disabled={isReadOnly || !!formData.parent_account}
                  min="1"
                  max="5"
                  className={`
                    w-full px-4 py-2.5 border rounded-lg
                    focus:ring-2 focus:ring-emerald-500 focus:border-transparent
                    disabled:bg-gray-50 disabled:text-gray-500
                    ${getFieldError('level') ? 'border-red-500' : 'border-gray-300'}
                  `}
                />
                {getFieldError('level') && (
                  <p className="mt-1 text-sm text-red-600 flex items-center gap-1">
                    <AlertCircle className="w-4 h-4" />
                    {getFieldError('level')}
                  </p>
                )}
                <p className="mt-1 text-xs text-gray-500">
                  {formData.parent_account ? 'Auto-calculated from parent' : 'Hierarchy level (1-5)'}
                </p>
              </div>
            </div>

            {/* Currency & Opening Balance */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Currency */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <DollarSign className="w-4 h-4 text-gray-500" />
                  Currency *
                </label>
                <select
                  value={formData.currency}
                  onChange={(e) => handleChange('currency', e.target.value)}
                  disabled={isReadOnly}
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-500"
                >
                  {currencyOptions.map((curr) => (
                    <option key={curr.value} value={curr.value}>
                      {curr.label}
                    </option>
                  ))}
                </select>
              </div>

              {/* Opening Balance */}
              <div>
                <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
                  <DollarSign className="w-4 h-4 text-gray-500" />
                  Opening Balance
                </label>
                <input
                  type="number"
                  value={formData.opening_balance}
                  onChange={(e) => handleChange('opening_balance', parseFloat(e.target.value) || 0)}
                  disabled={isReadOnly}
                  step="0.01"
                  className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-500"
                />
                <p className="mt-1 text-xs text-gray-500">
                  Initial balance for this account
                </p>
              </div>
            </div>

            {/* Account Settings */}
            <div className="bg-gray-50 rounded-lg p-4 space-y-4">
              <h3 className="text-sm font-semibold text-gray-900 flex items-center gap-2">
                <Info className="w-4 h-4 text-gray-500" />
                Account Settings
              </h3>

              {/* Control Account */}
              <label className="flex items-center justify-between p-3 bg-white rounded-lg border border-gray-200 cursor-pointer hover:bg-gray-50 transition-colors">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${formData.is_control_account ? 'bg-emerald-100' : 'bg-gray-100'}`}>
                    <Layers className={`w-5 h-5 ${formData.is_control_account ? 'text-emerald-600' : 'text-gray-400'}`} />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Control Account</p>
                    <p className="text-xs text-gray-500">Has child accounts (cannot post directly)</p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={formData.is_control_account}
                  onChange={(e) => !isReadOnly && handleChange('is_control_account', e.target.checked)}
                  disabled={isReadOnly}
                  className="w-5 h-5 text-emerald-600 rounded focus:ring-emerald-500"
                />
              </label>

              {/* Allow Posting */}
              <label className={`flex items-center justify-between p-3 bg-white rounded-lg border border-gray-200 cursor-pointer hover:bg-gray-50 transition-colors ${formData.is_control_account ? 'opacity-50 cursor-not-allowed' : ''}`}>
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${formData.allow_posting ? 'bg-blue-100' : 'bg-gray-100'}`}>
                    {formData.allow_posting ? (
                      <Unlock className="w-5 h-5 text-blue-600" />
                    ) : (
                      <Lock className="w-5 h-5 text-gray-400" />
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Allow Direct Posting</p>
                    <p className="text-xs text-gray-500">Enable journal entries to this account</p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={formData.allow_posting}
                  onChange={(e) => !isReadOnly && !formData.is_control_account && handleChange('allow_posting', e.target.checked)}
                  disabled={isReadOnly || formData.is_control_account}
                  className="w-5 h-5 text-blue-600 rounded focus:ring-blue-500"
                />
              </label>

              {/* Active Status */}
              <label className="flex items-center justify-between p-3 bg-white rounded-lg border border-gray-200 cursor-pointer hover:bg-gray-50 transition-colors">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${formData.is_active ? 'bg-green-100' : 'bg-gray-100'}`}>
                    <CheckCircle2 className={`w-5 h-5 ${formData.is_active ? 'text-green-600' : 'text-gray-400'}`} />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-900">Active Status</p>
                    <p className="text-xs text-gray-500">Account is active and visible</p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={formData.is_active}
                  onChange={(e) => !isReadOnly && handleChange('is_active', e.target.checked)}
                  disabled={isReadOnly}
                  className="w-5 h-5 text-green-600 rounded focus:ring-green-500"
                />
              </label>

              {/* Validation Warning */}
              {getFieldError('allow_posting') && (
                <div className="flex items-start gap-2 p-3 bg-orange-50 border border-orange-200 rounded-lg">
                  <AlertCircle className="w-5 h-5 text-orange-600 flex-shrink-0 mt-0.5" />
                  <p className="text-sm text-orange-800">{getFieldError('allow_posting')}</p>
                </div>
              )}
            </div>

            {/* Info Box */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <Info className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                <div className="text-sm text-blue-900">
                  <p className="font-medium mb-1">Account Guidelines:</p>
                  <ul className="list-disc list-inside space-y-1 text-blue-800">
                    <li>Account codes should follow your organization's numbering scheme</li>
                    <li>Control accounts are used for grouping and cannot have direct postings</li>
                    <li>Child accounts inherit the account type from their parent</li>
                    <li>Opening balance can be adjusted later through journal entries</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </form>

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
              <button
                onClick={handleSubmit}
                disabled={saving}
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
                    <span>{mode === 'edit' ? 'Update Account' : 'Create Account'}</span>
                  </>
                )}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
