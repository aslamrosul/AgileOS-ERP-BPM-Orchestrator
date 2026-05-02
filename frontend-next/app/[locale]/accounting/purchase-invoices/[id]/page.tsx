'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { 
  ArrowLeft, Edit, Trash2, CheckCircle, XCircle, Clock, FileText,
  Building, Calendar, DollarSign, User, Download, Printer, Send,
  AlertCircle, CreditCard, Eye, RefreshCw, MessageSquare
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
  tax_amount: number;
  discount_amount: number;
  paid_amount: number;
  status: 'draft' | 'submitted' | 'approved' | 'paid' | 'cancelled';
  payment_status: 'unpaid' | 'partial' | 'paid' | 'overdue';
  description: string;
  reference?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
  lines?: InvoiceLine[];
  payments?: Payment[];
}

interface InvoiceLine {
  id: string;
  description: string;
  account_id: string;
  account_code: string;
  account_name: string;
  quantity: number;
  unit_price: number;
  tax_rate: number;
  amount: number;
}

interface Payment {
  id: string;
  payment_number: string;
  payment_date: string;
  amount: number;
  payment_method: string;
  reference?: string;
  created_by: string;
  created_at: string;
}

export default function PurchaseInvoiceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const [invoice, setInvoice] = useState<PurchaseInvoice | null>(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

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
    } catch (error) {
      console.error('Failed to fetch invoice:', error);
      toast.error('Failed to load invoice details');
      router.push('/accounting/purchase-invoices');
    } finally {
      setLoading(false);
    }
  };

  const handleApprove = async () => {
    if (!confirm('Are you sure you want to approve this invoice?')) {
      return;
    }

    try {
      setActionLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${params.id}/approve`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to approve invoice');
      }

      toast.success('Invoice approved successfully');
      fetchInvoice();
    } catch (error) {
      console.error('Failed to approve invoice:', error);
      toast.error('Failed to approve invoice');
    } finally {
      setActionLoading(false);
    }
  };

  const handleCancel = async () => {
    if (!confirm('Are you sure you want to cancel this invoice? This action cannot be undone.')) {
      return;
    }

    try {
      setActionLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${params.id}/cancel`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to cancel invoice');
      }

      toast.success('Invoice cancelled successfully');
      fetchInvoice();
    } catch (error) {
      console.error('Failed to cancel invoice:', error);
      toast.error('Failed to cancel invoice');
    } finally {
      setActionLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this invoice? This action cannot be undone.')) {
      return;
    }

    try {
      setActionLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${params.id}`,
        { method: 'DELETE' }
      );

      if (!response.ok) {
        throw new Error('Failed to delete invoice');
      }

      toast.success('Invoice deleted successfully');
      router.push('/accounting/purchase-invoices');
    } catch (error) {
      console.error('Failed to delete invoice:', error);
      toast.error('Failed to delete invoice');
    } finally {
      setActionLoading(false);
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

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getStatusBadge = (status: string) => {
    const badges: Record<string, { bg: string; text: string; icon: any }> = {
      draft: { bg: 'bg-gray-100', text: 'text-gray-800', icon: Clock },
      submitted: { bg: 'bg-yellow-100', text: 'text-yellow-800', icon: AlertCircle },
      approved: { bg: 'bg-blue-100', text: 'text-blue-800', icon: CheckCircle },
      paid: { bg: 'bg-green-100', text: 'text-green-800', icon: CheckCircle },
      cancelled: { bg: 'bg-red-100', text: 'text-red-800', icon: XCircle }
    };
    const badge = badges[status] || { bg: 'bg-gray-100', text: 'text-gray-800', icon: FileText };
    const Icon = badge.icon;
    
    return (
      <span className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        <Icon className="w-4 h-4" />
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
  };

  const getPaymentStatusBadge = (status: string) => {
    const badges: Record<string, { bg: string; text: string; icon: any }> = {
      unpaid: { bg: 'bg-red-100', text: 'text-red-800', icon: XCircle },
      partial: { bg: 'bg-yellow-100', text: 'text-yellow-800', icon: Clock },
      paid: { bg: 'bg-green-100', text: 'text-green-800', icon: CheckCircle },
      overdue: { bg: 'bg-red-100', text: 'text-red-800', icon: AlertCircle }
    };
    const badge = badges[status] || { bg: 'bg-gray-100', text: 'text-gray-800', icon: DollarSign };
    const Icon = badge.icon;
    
    return (
      <span className={`inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        <Icon className="w-4 h-4" />
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
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

  const daysUntilDue = getDaysUntilDue(invoice.due_date);
  const isOverdue = daysUntilDue < 0 && invoice.payment_status !== 'paid';
  const outstandingAmount = invoice.total_amount - invoice.paid_amount;

  return (
    <div className="p-8">
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <Link href="/accounting/purchase-invoices" className="p-2 hover:bg-gray-100 rounded-lg transition-colors">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <div className="flex-1">
            <h1 className="text-3xl font-bold text-gray-900">{invoice.invoice_number}</h1>
            <p className="text-gray-600 mt-1">Purchase Invoice Details</p>
          </div>
          <div className="flex items-center gap-2">
            {getStatusBadge(invoice.status)}
            {getPaymentStatusBadge(invoice.payment_status)}
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button onClick={fetchInvoice}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
            <RefreshCw className="w-4 h-4" />
            Refresh
          </button>

          {invoice.status === 'draft' && (
            <Link href={`/accounting/purchase-invoices/${invoice.id}/edit`}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
              <Edit className="w-4 h-4" />
              Edit
            </Link>
          )}

          {invoice.status === 'submitted' && (
            <button onClick={handleApprove} disabled={actionLoading}
              className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50">
              <CheckCircle className="w-4 h-4" />
              Approve
            </button>
          )}

          {(invoice.status === 'approved' || invoice.payment_status === 'partial') && (
            <Link href={`/accounting/purchase-invoices/${invoice.id}/payment`}
              className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors">
              <CreditCard className="w-4 h-4" />
              Record Payment
            </Link>
          )}

          {invoice.status !== 'cancelled' && invoice.status !== 'paid' && (
            <button onClick={handleCancel} disabled={actionLoading}
              className="flex items-center gap-2 px-4 py-2 border border-red-300 text-red-600 rounded-lg hover:bg-red-50 transition-colors disabled:opacity-50">
              <XCircle className="w-4 h-4" />
              Cancel
            </button>
          )}

          {invoice.status === 'draft' && (
            <button onClick={handleDelete} disabled={actionLoading}
              className="flex items-center gap-2 px-4 py-2 border border-red-300 text-red-600 rounded-lg hover:bg-red-50 transition-colors disabled:opacity-50">
              <Trash2 className="w-4 h-4" />
              Delete
            </button>
          )}

          <button onClick={() => toast.info('Print feature coming soon')}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
            <Printer className="w-4 h-4" />
            Print
          </button>

          <button onClick={() => toast.info('Export feature coming soon')}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
            <Download className="w-4 h-4" />
            Export
          </button>
        </div>
      </div>

      {isOverdue && (
        <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-red-900">Invoice Overdue</p>
              <p className="text-sm text-red-700 mt-1">
                This invoice is {Math.abs(daysUntilDue)} days overdue. Outstanding amount: {formatCurrency(outstandingAmount)}
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Invoice Information</h2>
            
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-600 mb-1">Vendor</p>
                <div className="flex items-center gap-2">
                  <Building className="w-4 h-4 text-gray-400" />
                  <p className="font-medium text-gray-900">{invoice.vendor_name}</p>
                </div>
              </div>

              <div>
                <p className="text-sm text-gray-600 mb-1">Invoice Date</p>
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-gray-400" />
                  <p className="font-medium text-gray-900">{formatDate(invoice.invoice_date)}</p>
                </div>
              </div>

              <div>
                <p className="text-sm text-gray-600 mb-1">Due Date</p>
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-gray-400" />
                  <p className={`font-medium ${isOverdue ? 'text-red-600' : 'text-gray-900'}`}>
                    {formatDate(invoice.due_date)}
                  </p>
                </div>
                {invoice.payment_status !== 'paid' && (
                  <p className={`text-xs mt-1 ${isOverdue ? 'text-red-600' : 'text-gray-500'}`}>
                    {isOverdue ? `${Math.abs(daysUntilDue)} days overdue` : `${daysUntilDue} days left`}
                  </p>
                )}
              </div>

              {invoice.reference && (
                <div>
                  <p className="text-sm text-gray-600 mb-1">Reference</p>
                  <p className="font-medium text-gray-900">{invoice.reference}</p>
                </div>
              )}

              {invoice.description && (
                <div className="col-span-2">
                  <p className="text-sm text-gray-600 mb-1">Description</p>
                  <p className="text-gray-900">{invoice.description}</p>
                </div>
              )}
            </div>
          </div>

          {invoice.lines && invoice.lines.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
              <div className="p-6 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">Invoice Lines</h2>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Description</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Account</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Qty</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Unit Price</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Tax %</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Amount</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {invoice.lines.map((line) => (
                      <tr key={line.id}>
                        <td className="px-6 py-4">
                          <p className="text-sm font-medium text-gray-900">{line.description}</p>
                        </td>
                        <td className="px-6 py-4">
                          <p className="text-sm text-gray-900">{line.account_code}</p>
                          <p className="text-xs text-gray-500">{line.account_name}</p>
                        </td>
                        <td className="px-6 py-4 text-right text-sm text-gray-900">{line.quantity}</td>
                        <td className="px-6 py-4 text-right text-sm text-gray-900">{formatCurrency(line.unit_price)}</td>
                        <td className="px-6 py-4 text-right text-sm text-gray-900">{line.tax_rate}%</td>
                        <td className="px-6 py-4 text-right text-sm font-medium text-gray-900">{formatCurrency(line.amount)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {invoice.payments && invoice.payments.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
              <div className="p-6 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">Payment History</h2>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Payment #</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Method</th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Amount</th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Reference</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {invoice.payments.map((payment) => (
                      <tr key={payment.id}>
                        <td className="px-6 py-4 text-sm font-medium text-gray-900">{payment.payment_number}</td>
                        <td className="px-6 py-4 text-sm text-gray-900">{formatDate(payment.payment_date)}</td>
                        <td className="px-6 py-4 text-sm text-gray-900">{payment.payment_method}</td>
                        <td className="px-6 py-4 text-sm font-medium text-green-600 text-right">{formatCurrency(payment.amount)}</td>
                        <td className="px-6 py-4 text-sm text-gray-500">{payment.reference || '-'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>

        <div className="lg:col-span-1 space-y-6">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Financial Summary</h2>
            
            <div className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Subtotal:</span>
                <span className="font-medium text-gray-900">
                  {formatCurrency(invoice.total_amount - invoice.tax_amount)}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Tax Amount:</span>
                <span className="font-medium text-gray-900">{formatCurrency(invoice.tax_amount)}</span>
              </div>
              {invoice.discount_amount > 0 && (
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Discount:</span>
                  <span className="font-medium text-red-600">-{formatCurrency(invoice.discount_amount)}</span>
                </div>
              )}
              <div className="border-t border-gray-200 pt-3">
                <div className="flex justify-between mb-3">
                  <span className="font-semibold text-gray-900">Total Amount:</span>
                  <span className="text-lg font-bold text-gray-900">{formatCurrency(invoice.total_amount)}</span>
                </div>
              </div>
              {invoice.paid_amount > 0 && (
                <>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Paid Amount:</span>
                    <span className="font-medium text-green-600">{formatCurrency(invoice.paid_amount)}</span>
                  </div>
                  <div className="border-t border-gray-200 pt-3">
                    <div className="flex justify-between">
                      <span className="font-semibold text-gray-900">Outstanding:</span>
                      <span className={`text-lg font-bold ${outstandingAmount > 0 ? 'text-red-600' : 'text-green-600'}`}>
                        {formatCurrency(outstandingAmount)}
                      </span>
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Audit Trail</h2>
            
            <div className="space-y-3 text-sm">
              <div>
                <p className="text-gray-600 mb-1">Created By</p>
                <div className="flex items-center gap-2">
                  <User className="w-4 h-4 text-gray-400" />
                  <p className="font-medium text-gray-900">{invoice.created_by}</p>
                </div>
                <p className="text-xs text-gray-500 mt-1">{formatDateTime(invoice.created_at)}</p>
              </div>

              <div className="border-t border-gray-200 pt-3">
                <p className="text-gray-600 mb-1">Last Updated</p>
                <p className="text-xs text-gray-500">{formatDateTime(invoice.updated_at)}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
