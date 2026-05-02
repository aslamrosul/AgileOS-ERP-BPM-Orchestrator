'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { 
  ArrowLeft, Save, X, CreditCard, Calendar, DollarSign, 
  Building, FileText, AlertCircle, CheckCircle, Banknote,
  Hash, MessageSquare, Calculator, Info
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface PurchaseInvoice {
  id: string;
  invoice_number: string;
  vendor_id: string;
  vendor_name: string;
  invoice_date: string;
  due_date: string;
  total_amount: number;
  paid_amount: number;
  status: string;
  payment_status: string;
}

export default function RecordPaymentPage() {
  const params = useParams();
  const router = useRouter();
  const [invoice, setInvoice] = useState<PurchaseInvoice | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Form data
  const [paymentDate, setPaymentDate] = useState(new Date().toISOString().split('T')[0]);
  const [paymentAmount, setPaymentAmount] = useState<number>(0);
  const [paymentMethod, setPaymentMethod] = useState<'cash' | 'bank_transfer' | 'check' | 'credit_card'>('bank_transfer');
  const [bankAccount, setBankAccount] = useState('');
  const [referenceNumber, setReferenceNumber] = useState('');
  const [notes, setNotes] = useState('');

  useEffect(() => {
    fetchInvoice();
  }, [params.id]);

  const fetchInvoice = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${params.id}`
      );

      if (!response.ok) {
        throw new Error('Failed to fetch invoice');
      }

      const data = await response.json();
      setInvoice(data);
      
      // Set default payment amount to outstanding balance
      const outstanding = data.total_amount - data.paid_amount;
      setPaymentAmount(outstanding);
    } catch (error) {
      console.error('Failed to fetch invoice:', error);
      toast.error('Failed to load invoice details');
      router.push('/accounting/purchase-invoices');
    } finally {
      setLoading(false);
    }
  };

  const handlePayFull = () => {
    if (invoice) {
      const outstanding = invoice.total_amount - invoice.paid_amount;
      setPaymentAmount(outstanding);
    }
  };

  const validateForm = (): boolean => {
    if (!paymentDate) {
      toast.error('Payment date is required');
      return false;
    }

    const selectedDate = new Date(paymentDate);
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    if (selectedDate > today) {
      toast.error('Payment date cannot be in the future');
      return false;
    }

    if (!paymentAmount || paymentAmount <= 0) {
      toast.error('Payment amount must be greater than 0');
      return false;
    }

    if (invoice) {
      const outstanding = invoice.total_amount - invoice.paid_amount;
      if (paymentAmount > outstanding) {
        toast.error(`Payment amount cannot exceed outstanding balance (${formatCurrency(outstanding)})`);
        return false;
      }
    }

    if (paymentMethod === 'bank_transfer' && !bankAccount.trim()) {
      toast.error('Bank account is required for bank transfer');
      return false;
    }

    if (paymentMethod === 'check' && !referenceNumber.trim()) {
      toast.error('Check number is required for check payment');
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
        invoice_id: params.id,
        payment_date: new Date(paymentDate).toISOString(),
        amount: paymentAmount,
        payment_method: paymentMethod,
        bank_account: bankAccount || undefined,
        reference_number: referenceNumber || undefined,
        notes: notes || undefined
      };

      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${params.id}/payments`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        }
      );

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to record payment');
      }

      toast.success('Payment recorded successfully');
      router.push(`/accounting/purchase-invoices/${params.id}`);
    } catch (error: any) {
      console.error('Failed to record payment:', error);
      toast.error(error.message || 'Failed to record payment');
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const getDaysUntilDue = (dueDate: string) => {
    const today = new Date();
    const due = new Date(dueDate);
    const diffTime = due.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading invoice details...</p>
        </div>
      </div>
    );
  }

  if (!invoice) {
    return null;
  }

  const outstandingAmount = invoice.total_amount - invoice.paid_amount;
  const remainingAfterPayment = outstandingAmount - paymentAmount;
  const isPartialPayment = paymentAmount < outstandingAmount && paymentAmount > 0;
  const daysUntilDue = getDaysUntilDue(invoice.due_date);
  const isOverdue = daysUntilDue < 0;

  return (
    <div className="p-8">
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <Link href={`/accounting/purchase-invoices/${params.id}`} 
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Record Payment</h1>
            <p className="text-gray-600 mt-1">Record payment for invoice {invoice.invoice_number}</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Invoice Summary */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <FileText className="w-5 h-5 text-emerald-600" />
                Invoice Summary
              </h2>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-sm text-gray-600 mb-1">Invoice Number</p>
                  <p className="font-medium text-gray-900">{invoice.invoice_number}</p>
                </div>

                <div>
                  <p className="text-sm text-gray-600 mb-1">Vendor</p>
                  <div className="flex items-center gap-2">
                    <Building className="w-4 h-4 text-gray-400" />
                    <p className="font-medium text-gray-900">{invoice.vendor_name}</p>
                  </div>
                </div>

                <div>
                  <p className="text-sm text-gray-600 mb-1">Invoice Date</p>
                  <p className="text-gray-900">{formatDate(invoice.invoice_date)}</p>
                </div>

                <div>
                  <p className="text-sm text-gray-600 mb-1">Due Date</p>
                  <p className={`font-medium ${isOverdue ? 'text-red-600' : 'text-gray-900'}`}>
                    {formatDate(invoice.due_date)}
                  </p>
                  {isOverdue && (
                    <p className="text-xs text-red-600 mt-1">
                      {Math.abs(daysUntilDue)} days overdue
                    </p>
                  )}
                </div>

                <div>
                  <p className="text-sm text-gray-600 mb-1">Total Amount</p>
                  <p className="text-lg font-bold text-gray-900">{formatCurrency(invoice.total_amount)}</p>
                </div>

                <div>
                  <p className="text-sm text-gray-600 mb-1">Paid Amount</p>
                  <p className="text-lg font-semibold text-green-600">{formatCurrency(invoice.paid_amount)}</p>
                </div>

                <div className="md:col-span-2 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <p className="text-sm text-gray-600 mb-1">Outstanding Balance</p>
                  <p className="text-2xl font-bold text-yellow-700">{formatCurrency(outstandingAmount)}</p>
                </div>
              </div>
            </div>

            {/* Payment Information */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <CreditCard className="w-5 h-5 text-emerald-600" />
                Payment Information
              </h2>

              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Payment Date <span className="text-red-500">*</span>
                    </label>
                    <div className="relative">
                      <Calendar className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                      <input
                        type="date"
                        value={paymentDate}
                        onChange={(e) => setPaymentDate(e.target.value)}
                        max={new Date().toISOString().split('T')[0]}
                        required
                        className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Payment Amount (IDR) <span className="text-red-500">*</span>
                    </label>
                    <div className="relative">
                      <DollarSign className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                      <input
                        type="number"
                        value={paymentAmount}
                        onChange={(e) => setPaymentAmount(parseFloat(e.target.value) || 0)}
                        min="0"
                        max={outstandingAmount}
                        step="0.01"
                        required
                        className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                      />
                    </div>
                    <button
                      type="button"
                      onClick={handlePayFull}
                      className="mt-2 text-sm text-emerald-600 hover:text-emerald-700 font-medium"
                    >
                      Pay full amount ({formatCurrency(outstandingAmount)})
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Payment Method <span className="text-red-500">*</span>
                  </label>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                    {[
                      { value: 'bank_transfer', label: 'Bank Transfer', icon: Banknote },
                      { value: 'cash', label: 'Cash', icon: DollarSign },
                      { value: 'check', label: 'Check', icon: FileText },
                      { value: 'credit_card', label: 'Credit Card', icon: CreditCard }
                    ].map((method) => {
                      const Icon = method.icon;
                      return (
                        <button
                          key={method.value}
                          type="button"
                          onClick={() => setPaymentMethod(method.value as any)}
                          className={`flex flex-col items-center gap-2 p-4 border-2 rounded-lg transition-all ${
                            paymentMethod === method.value
                              ? 'border-emerald-600 bg-emerald-50 text-emerald-700'
                              : 'border-gray-200 hover:border-gray-300 text-gray-700'
                          }`}
                        >
                          <Icon className="w-6 h-6" />
                          <span className="text-sm font-medium">{method.label}</span>
                        </button>
                      );
                    })}
                  </div>
                </div>

                {paymentMethod === 'bank_transfer' && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Bank Account <span className="text-red-500">*</span>
                    </label>
                    <input
                      type="text"
                      value={bankAccount}
                      onChange={(e) => setBankAccount(e.target.value)}
                      placeholder="Enter bank account number or name"
                      required
                      className="w-full px-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                    />
                  </div>
                )}

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Reference Number {paymentMethod === 'check' && <span className="text-red-500">*</span>}
                  </label>
                  <div className="relative">
                    <Hash className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                    <input
                      type="text"
                      value={referenceNumber}
                      onChange={(e) => setReferenceNumber(e.target.value)}
                      placeholder={
                        paymentMethod === 'check' 
                          ? 'Check number' 
                          : paymentMethod === 'bank_transfer'
                          ? 'Transfer reference'
                          : 'Payment reference'
                      }
                      required={paymentMethod === 'check'}
                      className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Notes
                  </label>
                  <div className="relative">
                    <MessageSquare className="absolute left-3 top-3 w-4 h-4 text-gray-400" />
                    <textarea
                      value={notes}
                      onChange={(e) => setNotes(e.target.value)}
                      placeholder="Additional notes about this payment..."
                      rows={3}
                      className="w-full pl-10 pr-4 py-2.5 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent resize-none"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Partial Payment Warning */}
            {isPartialPayment && (
              <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <div className="flex items-start gap-3">
                  <AlertCircle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <p className="text-sm font-medium text-yellow-900">Partial Payment</p>
                    <p className="text-sm text-yellow-700 mt-1">
                      This is a partial payment. Remaining balance after this payment: {formatCurrency(remainingAfterPayment)}
                    </p>
                  </div>
                </div>
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex items-center gap-3">
              <button
                type="submit"
                disabled={saving}
                className="flex items-center gap-2 px-6 py-3 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {saving ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                    <span>Recording Payment...</span>
                  </>
                ) : (
                  <>
                    <Save className="w-4 h-4" />
                    <span>Record Payment</span>
                  </>
                )}
              </button>

              <Link
                href={`/accounting/purchase-invoices/${params.id}`}
                className="flex items-center gap-2 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors font-medium"
              >
                <X className="w-4 h-4" />
                Cancel
              </Link>
            </div>
          </form>
        </div>

        {/* Sidebar */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 sticky top-8 space-y-6">
            <div>
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <Calculator className="w-5 h-5 text-emerald-600" />
                Payment Summary
              </h2>

              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Outstanding:</span>
                  <span className="font-medium text-gray-900">{formatCurrency(outstandingAmount)}</span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Payment Amount:</span>
                  <span className="font-semibold text-emerald-600">{formatCurrency(paymentAmount)}</span>
                </div>
                <div className="border-t border-gray-200 pt-3">
                  <div className="flex justify-between">
                    <span className="font-semibold text-gray-900">Remaining Balance:</span>
                    <span className={`text-lg font-bold ${remainingAfterPayment > 0 ? 'text-yellow-600' : 'text-green-600'}`}>
                      {formatCurrency(remainingAfterPayment)}
                    </span>
                  </div>
                </div>
              </div>

              {remainingAfterPayment === 0 && (
                <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-lg">
                  <div className="flex items-center gap-2 text-green-700">
                    <CheckCircle className="w-5 h-5" />
                    <span className="text-sm font-medium">Full Payment</span>
                  </div>
                  <p className="text-xs text-green-600 mt-1">
                    This payment will fully settle the invoice
                  </p>
                </div>
              )}
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex items-start gap-3">
                <Info className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="text-xs font-medium text-blue-900 mb-2">Payment Guidelines</p>
                  <ul className="text-xs text-blue-700 space-y-1">
                    <li>• Payment date cannot be in the future</li>
                    <li>• Amount must not exceed outstanding balance</li>
                    <li>• Partial payments are allowed</li>
                    <li>• Bank account required for transfers</li>
                    <li>• Check number required for check payments</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
