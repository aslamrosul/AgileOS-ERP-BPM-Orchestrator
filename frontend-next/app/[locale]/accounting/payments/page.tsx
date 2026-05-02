'use client';

import { useEffect, useState } from 'react';
import { 
  Plus, Search, Filter, Download, Eye, Edit, Trash2, CheckCircle,
  Clock, XCircle, AlertCircle, DollarSign, Calendar, User, Building,
  CreditCard, Banknote, FileText, RefreshCw, TrendingUp, TrendingDown,
  ArrowUpRight, ArrowDownRight
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface Payment {
  id: string;
  payment_number: string;
  payment_type: 'vendor_payment' | 'customer_receipt';
  party_id: string;
  party_name: string;
  payment_date: string;
  payment_method: 'cash' | 'bank_transfer' | 'check' | 'credit_card';
  amount: number;
  bank_account: string;
  reference_number: string;
  status: 'draft' | 'submitted' | 'cleared' | 'cancelled';
  description: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

interface Vendor {
  id: string;
  vendor_code: string;
  vendor_name: string;
}

interface Customer {
  id: string;
  customer_code: string;
  customer_name: string;
}

export default function PaymentsPage() {
  const [payments, setPayments] = useState<Payment[]>([]);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [filteredPayments, setFilteredPayments] = useState<Payment[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [paymentTypeFilter, setPaymentTypeFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [paymentMethodFilter, setPaymentMethodFilter] = useState<string>('all');
  const [partyFilter, setPartyFilter] = useState<string>('all');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    filterPayments();
  }, [payments, searchTerm, paymentTypeFilter, statusFilter, paymentMethodFilter, partyFilter, dateFrom, dateTo]);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      const [paymentsRes, vendorsRes, customersRes] = await Promise.all([
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/payments`),
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/vendors`),
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/customers`)
      ]);

      if (!paymentsRes.ok || !vendorsRes.ok || !customersRes.ok) {
        throw new Error('Failed to fetch data');
      }

      const [paymentsData, vendorsData, customersData] = await Promise.all([
        paymentsRes.json(),
        vendorsRes.json(),
        customersRes.json()
      ]);

      setPayments(paymentsData || []);
      setVendors(vendorsData || []);
      setCustomers(customersData || []);
    } catch (error) {
      console.error('Failed to fetch data:', error);
      toast.error('Failed to load payments');
    } finally {
      setLoading(false);
    }
  };

  const filterPayments = () => {
    let filtered = [...payments];

    if (searchTerm) {
      filtered = filtered.filter(payment =>
        payment.payment_number.toLowerCase().includes(searchTerm.toLowerCase()) ||
        payment.party_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        payment.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (payment.reference_number && payment.reference_number.toLowerCase().includes(searchTerm.toLowerCase()))
      );
    }

    if (paymentTypeFilter !== 'all') {
      filtered = filtered.filter(payment => payment.payment_type === paymentTypeFilter);
    }

    if (statusFilter !== 'all') {
      filtered = filtered.filter(payment => payment.status === statusFilter);
    }

    if (paymentMethodFilter !== 'all') {
      filtered = filtered.filter(payment => payment.payment_method === paymentMethodFilter);
    }

    if (partyFilter !== 'all') {
      filtered = filtered.filter(payment => payment.party_id === partyFilter);
    }

    if (dateFrom) {
      filtered = filtered.filter(payment => 
        new Date(payment.payment_date) >= new Date(dateFrom)
      );
    }
    if (dateTo) {
      filtered = filtered.filter(payment => 
        new Date(payment.payment_date) <= new Date(dateTo)
      );
    }

    setFilteredPayments(filtered);
  };

  const handleClearPayment = async (paymentId: string) => {
    if (!confirm('Are you sure you want to clear this payment?')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/payments/${paymentId}/clear`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to clear payment');
      }

      toast.success('Payment cleared successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to clear payment:', error);
      toast.error('Failed to clear payment');
    }
  };

  const handleCancelPayment = async (paymentId: string) => {
    if (!confirm('Are you sure you want to cancel this payment? This action cannot be undone.')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/payments/${paymentId}/cancel`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to cancel payment');
      }

      toast.success('Payment cancelled successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to cancel payment:', error);
      toast.error('Failed to cancel payment');
    }
  };

  const handleDeletePayment = async (paymentId: string) => {
    if (!confirm('Are you sure you want to delete this payment? This action cannot be undone.')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/payments/${paymentId}`,
        { method: 'DELETE' }
      );

      if (!response.ok) {
        throw new Error('Failed to delete payment');
      }

      toast.success('Payment deleted successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to delete payment:', error);
      toast.error('Failed to delete payment');
    }
  };

  const getStatusBadge = (status: string) => {
    const badges: Record<string, { bg: string; text: string; icon: any }> = {
      draft: { bg: 'bg-gray-100', text: 'text-gray-800', icon: Clock },
      submitted: { bg: 'bg-yellow-100', text: 'text-yellow-800', icon: AlertCircle },
      cleared: { bg: 'bg-green-100', text: 'text-green-800', icon: CheckCircle },
      cancelled: { bg: 'bg-red-100', text: 'text-red-800', icon: XCircle }
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

  const getPaymentMethodBadge = (method: string) => {
    const badges: Record<string, { bg: string; text: string; icon: any }> = {
      cash: { bg: 'bg-green-100', text: 'text-green-800', icon: Banknote },
      bank_transfer: { bg: 'bg-blue-100', text: 'text-blue-800', icon: Building },
      check: { bg: 'bg-purple-100', text: 'text-purple-800', icon: FileText },
      credit_card: { bg: 'bg-orange-100', text: 'text-orange-800', icon: CreditCard }
    };
    const badge = badges[method] || { bg: 'bg-gray-100', text: 'text-gray-800', icon: DollarSign };
    const Icon = badge.icon;
    
    const label = method.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
    
    return (
      <span className={`inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        <Icon className="w-3 h-3" />
        {label}
      </span>
    );
  };

  const getPaymentTypeIcon = (type: string) => {
    if (type === 'vendor_payment') {
      return <ArrowUpRight className="w-4 h-4 text-red-600" />;
    }
    return <ArrowDownRight className="w-4 h-4 text-green-600" />;
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
          <p className="text-gray-600">Loading payments...</p>
        </div>
      </div>
    );
  }

  const stats = {
    total: payments.length,
    vendorPayments: payments.filter(p => p.payment_type === 'vendor_payment').length,
    customerReceipts: payments.filter(p => p.payment_type === 'customer_receipt').length,
    draft: payments.filter(p => p.status === 'draft').length,
    submitted: payments.filter(p => p.status === 'submitted').length,
    cleared: payments.filter(p => p.status === 'cleared').length,
    totalVendorPayments: payments.filter(p => p.payment_type === 'vendor_payment').reduce((sum, p) => sum + p.amount, 0),
    totalCustomerReceipts: payments.filter(p => p.payment_type === 'customer_receipt').reduce((sum, p) => sum + p.amount, 0),
    netCashFlow: payments.filter(p => p.payment_type === 'customer_receipt').reduce((sum, p) => sum + p.amount, 0) - 
                 payments.filter(p => p.payment_type === 'vendor_payment').reduce((sum, p) => sum + p.amount, 0)
  };

  return (
    <div className="p-8">
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Payments</h1>
            <p className="text-gray-600 mt-2">Manage vendor payments and customer receipts</p>
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
              href="/accounting/payments/new"
              className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              <span>New Payment</span>
            </Link>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Total Payments</p>
          <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
        </div>
        <div className="bg-red-50 rounded-lg shadow-sm p-4 border border-red-200">
          <p className="text-xs text-red-700 mb-1">Vendor Payments</p>
          <p className="text-2xl font-bold text-red-900">{stats.vendorPayments}</p>
        </div>
        <div className="bg-green-50 rounded-lg shadow-sm p-4 border border-green-200">
          <p className="text-xs text-green-700 mb-1">Customer Receipts</p>
          <p className="text-2xl font-bold text-green-900">{stats.customerReceipts}</p>
        </div>
        <div className="bg-yellow-50 rounded-lg shadow-sm p-4 border border-yellow-200">
          <p className="text-xs text-yellow-700 mb-1">Pending</p>
          <p className="text-2xl font-bold text-yellow-900">{stats.submitted}</p>
        </div>
        <div className="bg-blue-50 rounded-lg shadow-sm p-4 border border-blue-200">
          <p className="text-xs text-blue-700 mb-1">Cleared</p>
          <p className="text-2xl font-bold text-blue-900">{stats.cleared}</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-gradient-to-br from-red-500 to-red-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Total Vendor Payments</p>
            <ArrowUpRight className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalVendorPayments)}</p>
          <p className="text-xs opacity-75 mt-2">Money out</p>
        </div>
        <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Total Customer Receipts</p>
            <ArrowDownRight className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalCustomerReceipts)}</p>
          <p className="text-xs opacity-75 mt-2">Money in</p>
        </div>
        <div className={`bg-gradient-to-br ${stats.netCashFlow >= 0 ? 'from-blue-500 to-blue-600' : 'from-orange-500 to-orange-600'} rounded-lg shadow-lg p-6 text-white`}>
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Net Cash Flow</p>
            {stats.netCashFlow >= 0 ? <TrendingUp className="w-5 h-5 opacity-75" /> : <TrendingDown className="w-5 h-5 opacity-75" />}
          </div>
          <p className="text-3xl font-bold">{formatCurrency(Math.abs(stats.netCashFlow))}</p>
          <p className="text-xs opacity-75 mt-2">{stats.netCashFlow >= 0 ? 'Positive' : 'Negative'}</p>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm p-6 border border-gray-200 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          <div className="md:col-span-2">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search by payment number, party, or description..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>
          </div>

          <div>
            <select
              value={paymentTypeFilter}
              onChange={(e) => setPaymentTypeFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Types</option>
              <option value="vendor_payment">Vendor Payment</option>
              <option value="customer_receipt">Customer Receipt</option>
            </select>
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
              <option value="cleared">Cleared</option>
              <option value="cancelled">Cancelled</option>
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

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Payment Method</label>
            <select
              value={paymentMethodFilter}
              onChange={(e) => setPaymentMethodFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Methods</option>
              <option value="cash">Cash</option>
              <option value="bank_transfer">Bank Transfer</option>
              <option value="check">Check</option>
              <option value="credit_card">Credit Card</option>
            </select>
          </div>

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

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Payment Number
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Type
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Party
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Payment Date
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Method
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
              {filteredPayments.map((payment) => {
                const isVendorPayment = payment.payment_type === 'vendor_payment';
                
                return (
                  <tr key={payment.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <FileText className="w-4 h-4 text-gray-400" />
                        <div>
                          <p className="font-mono text-sm font-medium text-gray-900">
                            {payment.payment_number}
                          </p>
                          {payment.reference_number && (
                            <p className="text-xs text-gray-500">Ref: {payment.reference_number}</p>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        {getPaymentTypeIcon(payment.payment_type)}
                        <span className={`text-sm font-medium ${isVendorPayment ? 'text-red-600' : 'text-green-600'}`}>
                          {isVendorPayment ? 'Payment' : 'Receipt'}
                        </span>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        {isVendorPayment ? <Building className="w-4 h-4 text-gray-400" /> : <User className="w-4 h-4 text-gray-400" />}
                        <div>
                          <p className="text-sm font-medium text-gray-900">{payment.party_name}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2 text-sm text-gray-900">
                        <Calendar className="w-4 h-4 text-gray-400" />
                        {formatDate(payment.payment_date)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getPaymentMethodBadge(payment.payment_method)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(payment.status)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <div className="text-sm">
                        <p className={`font-medium ${isVendorPayment ? 'text-red-600' : 'text-green-600'}`}>
                          {formatCurrency(payment.amount)}
                        </p>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Link
                          href={`/accounting/payments/${payment.id}`}
                          className="p-1.5 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          title="View Details"
                        >
                          <Eye className="w-4 h-4" />
                        </Link>
                        {payment.status === 'draft' && (
                          <>
                            <Link
                              href={`/accounting/payments/${payment.id}/edit`}
                              className="p-1.5 text-gray-600 hover:text-emerald-600 hover:bg-emerald-50 rounded transition-colors"
                              title="Edit Payment"
                            >
                              <Edit className="w-4 h-4" />
                            </Link>
                            <button
                              onClick={() => handleDeletePayment(payment.id)}
                              className="p-1.5 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                              title="Delete Payment"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </>
                        )}
                        {payment.status === 'submitted' && (
                          <button
                            onClick={() => handleClearPayment(payment.id)}
                            className="p-1.5 text-gray-600 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
                            title="Clear Payment"
                          >
                            <CheckCircle className="w-4 h-4" />
                          </button>
                        )}
                        {payment.status !== 'cleared' && payment.status !== 'cancelled' && (
                          <button
                            onClick={() => handleCancelPayment(payment.id)}
                            className="p-1.5 text-gray-600 hover:text-orange-600 hover:bg-orange-50 rounded transition-colors"
                            title="Cancel Payment"
                          >
                            <XCircle className="w-4 h-4" />
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

        {filteredPayments.length === 0 && (
          <div className="text-center py-12">
            <DollarSign className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">No payments found</p>
            <p className="text-sm text-gray-500 mt-2">
              {searchTerm || paymentTypeFilter !== 'all' || statusFilter !== 'all'
                ? 'Try adjusting your filters'
                : 'Create your first payment to get started'
              }
            </p>
            <Link
              href="/accounting/payments/new"
              className="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-4 h-4" />
              <span>Create Payment</span>
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}
