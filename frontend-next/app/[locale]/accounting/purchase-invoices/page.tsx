'use client';

import { useEffect, useState } from 'react';
import { 
  Plus, Search, Filter, Download, Upload, Eye, Edit, Trash2, CheckCircle,
  Clock, XCircle, AlertCircle, FileText, Calendar, DollarSign, User,
  TrendingUp, TrendingDown, RefreshCw, Building, CreditCard, ArrowUpRight,
  Percent, Package, ShoppingCart, Truck, Receipt
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
}

interface Vendor {
  id: string;
  vendor_code: string;
  vendor_name: string;
  vendor_type: string;
  current_balance: number;
  is_active: boolean;
}

export default function PurchaseInvoicesPage() {
  const [invoices, setInvoices] = useState<PurchaseInvoice[]>([]);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [filteredInvoices, setFilteredInvoices] = useState<PurchaseInvoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [paymentStatusFilter, setPaymentStatusFilter] = useState<string>('all');
  const [vendorFilter, setVendorFilter] = useState<string>('all');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  useEffect(() => {
    fetchData();
  }, []);

  useEffect(() => {
    filterInvoices();
  }, [invoices, searchTerm, statusFilter, paymentStatusFilter, vendorFilter, dateFrom, dateTo]);

  const fetchData = async () => {
    try {
      setLoading(true);
      
      // Fetch invoices and vendors in parallel
      const [invoicesRes, vendorsRes] = await Promise.all([
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices`),
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/vendors`)
      ]);

      if (!invoicesRes.ok || !vendorsRes.ok) {
        throw new Error('Failed to fetch data');
      }

      const [invoicesData, vendorsData] = await Promise.all([
        invoicesRes.json(),
        vendorsRes.json()
      ]);

      setInvoices(invoicesData || []);
      setVendors(vendorsData || []);
    } catch (error) {
      console.error('Failed to fetch data:', error);
      toast.error('Failed to load purchase invoices');
    } finally {
      setLoading(false);
    }
  };

  const filterInvoices = () => {
    let filtered = [...invoices];

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(invoice =>
        invoice.invoice_number.toLowerCase().includes(searchTerm.toLowerCase()) ||
        invoice.vendor_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        invoice.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (invoice.reference && invoice.reference.toLowerCase().includes(searchTerm.toLowerCase()))
      );
    }

    // Status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(invoice => invoice.status === statusFilter);
    }

    // Payment status filter
    if (paymentStatusFilter !== 'all') {
      filtered = filtered.filter(invoice => invoice.payment_status === paymentStatusFilter);
    }

    // Vendor filter
    if (vendorFilter !== 'all') {
      filtered = filtered.filter(invoice => invoice.vendor_id === vendorFilter);
    }

    // Date range filter
    if (dateFrom) {
      filtered = filtered.filter(invoice => 
        new Date(invoice.invoice_date) >= new Date(dateFrom)
      );
    }
    if (dateTo) {
      filtered = filtered.filter(invoice => 
        new Date(invoice.invoice_date) <= new Date(dateTo)
      );
    }

    setFilteredInvoices(filtered);
  };

  const handleApproveInvoice = async (invoiceId: string) => {
    if (!confirm('Are you sure you want to approve this purchase invoice?')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${invoiceId}/approve`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to approve invoice');
      }

      toast.success('Purchase invoice approved successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to approve invoice:', error);
      toast.error('Failed to approve purchase invoice');
    }
  };

  const handleCancelInvoice = async (invoiceId: string) => {
    if (!confirm('Are you sure you want to cancel this invoice? This action cannot be undone.')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${invoiceId}/cancel`,
        { method: 'POST' }
      );

      if (!response.ok) {
        throw new Error('Failed to cancel invoice');
      }

      toast.success('Invoice cancelled successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to cancel invoice:', error);
      toast.error('Failed to cancel invoice');
    }
  };

  const handleDeleteInvoice = async (invoiceId: string) => {
    if (!confirm('Are you sure you want to delete this invoice? This action cannot be undone.')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/purchase-invoices/${invoiceId}`,
        { method: 'DELETE' }
      );

      if (!response.ok) {
        throw new Error('Failed to delete invoice');
      }

      toast.success('Invoice deleted successfully');
      fetchData();
    } catch (error) {
      console.error('Failed to delete invoice:', error);
      toast.error('Failed to delete invoice');
    }
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
      <span className={`inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        <Icon className="w-3 h-3" />
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
      <span className={`inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        <Icon className="w-3 h-3" />
        {status.charAt(0).toUpperCase() + status.slice(1)}
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
          <p className="text-gray-600">Loading purchase invoices...</p>
        </div>
      </div>
    );
  }

  const stats = {
    total: invoices.length,
    draft: invoices.filter(i => i.status === 'draft').length,
    submitted: invoices.filter(i => i.status === 'submitted').length,
    approved: invoices.filter(i => i.status === 'approved').length,
    paid: invoices.filter(i => i.status === 'paid').length,
    unpaid: invoices.filter(i => i.payment_status === 'unpaid').length,
    overdue: invoices.filter(i => i.payment_status === 'overdue').length,
    totalAmount: invoices.reduce((sum, i) => sum + i.total_amount, 0),
    totalPaid: invoices.reduce((sum, i) => sum + i.paid_amount, 0),
    totalOutstanding: invoices.reduce((sum, i) => sum + (i.total_amount - i.paid_amount), 0)
  };

  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Purchase Invoices</h1>
            <p className="text-gray-600 mt-2">Manage vendor invoices and account payable</p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/accounting/vendors"
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Building className="w-4 h-4" />
              <span>Vendors</span>
            </Link>
            <button
              onClick={() => toast.info('Export feature coming soon')}
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Download className="w-4 h-4" />
              <span>Export</span>
            </button>
            <Link
              href="/accounting/purchase-invoices/new"
              className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              <span>New Invoice</span>
            </Link>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Total Invoices</p>
          <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
        </div>
        <div className="bg-yellow-50 rounded-lg shadow-sm p-4 border border-yellow-200">
          <p className="text-xs text-yellow-700 mb-1">Pending Approval</p>
          <p className="text-2xl font-bold text-yellow-900">{stats.submitted}</p>
        </div>
        <div className="bg-blue-50 rounded-lg shadow-sm p-4 border border-blue-200">
          <p className="text-xs text-blue-700 mb-1">Approved</p>
          <p className="text-2xl font-bold text-blue-900">{stats.approved}</p>
        </div>
        <div className="bg-red-50 rounded-lg shadow-sm p-4 border border-red-200">
          <p className="text-xs text-red-700 mb-1">Unpaid</p>
          <p className="text-2xl font-bold text-red-900">{stats.unpaid}</p>
        </div>
        <div className="bg-green-50 rounded-lg shadow-sm p-4 border border-green-200">
          <p className="text-xs text-green-700 mb-1">Paid</p>
          <p className="text-2xl font-bold text-green-900">{stats.paid}</p>
        </div>
      </div>

      {/* Financial Summary */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Total Invoice Amount</p>
            <DollarSign className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalAmount)}</p>
          <p className="text-xs opacity-75 mt-2">All invoices combined</p>
        </div>
        <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Total Paid</p>
            <CheckCircle className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalPaid)}</p>
          <p className="text-xs opacity-75 mt-2">Payments made</p>
        </div>
        <div className="bg-gradient-to-br from-red-500 to-red-600 rounded-lg shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <p className="text-sm font-medium opacity-90">Outstanding Balance</p>
            <AlertCircle className="w-5 h-5 opacity-75" />
          </div>
          <p className="text-3xl font-bold">{formatCurrency(stats.totalOutstanding)}</p>
          <p className="text-xs opacity-75 mt-2">{stats.overdue} overdue invoices</p>
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
                placeholder="Search by invoice number, vendor, or description..."
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
              <option value="submitted">Submitted</option>
              <option value="approved">Approved</option>
              <option value="paid">Paid</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>

          {/* Payment Status Filter */}
          <div>
            <select
              value={paymentStatusFilter}
              onChange={(e) => setPaymentStatusFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Payment Status</option>
              <option value="unpaid">Unpaid</option>
              <option value="partial">Partial</option>
              <option value="paid">Paid</option>
              <option value="overdue">Overdue</option>
            </select>
          </div>

          {/* Refresh */}
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

        {/* Additional Filters */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
          {/* Vendor Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Vendor</label>
            <select
              value={vendorFilter}
              onChange={(e) => setVendorFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Vendors</option>
              {vendors.map(vendor => (
                <option key={vendor.id} value={vendor.id}>
                  {vendor.vendor_name} ({vendor.vendor_code})
                </option>
              ))}
            </select>
          </div>

          {/* Date From */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">From Date</label>
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            />
          </div>

          {/* Date To */}
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

      {/* Invoices Table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Invoice Number
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Vendor
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Invoice Date
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Due Date
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Payment
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
              {filteredInvoices.map((invoice) => {
                const daysUntilDue = getDaysUntilDue(invoice.due_date);
                const isOverdue = daysUntilDue < 0 && invoice.payment_status !== 'paid';
                
                return (
                  <tr key={invoice.id} className={`hover:bg-gray-50 transition-colors ${isOverdue ? 'bg-red-50' : ''}`}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <FileText className="w-4 h-4 text-gray-400" />
                        <div>
                          <p className="font-mono text-sm font-medium text-gray-900">
                            {invoice.invoice_number}
                          </p>
                          {invoice.reference && (
                            <p className="text-xs text-gray-500">Ref: {invoice.reference}</p>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Building className="w-4 h-4 text-gray-400" />
                        <div>
                          <p className="text-sm font-medium text-gray-900">{invoice.vendor_name}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2 text-sm text-gray-900">
                        <Calendar className="w-4 h-4 text-gray-400" />
                        {formatDate(invoice.invoice_date)}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm">
                        <p className={`font-medium ${isOverdue ? 'text-red-600' : 'text-gray-900'}`}>
                          {formatDate(invoice.due_date)}
                        </p>
                        {invoice.payment_status !== 'paid' && (
                          <p className={`text-xs ${isOverdue ? 'text-red-600' : 'text-gray-500'}`}>
                            {isOverdue ? `${Math.abs(daysUntilDue)} days overdue` : `${daysUntilDue} days left`}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getStatusBadge(invoice.status)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {getPaymentStatusBadge(invoice.payment_status)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <div className="text-sm">
                        <p className="font-medium text-gray-900">{formatCurrency(invoice.total_amount)}</p>
                        {invoice.paid_amount > 0 && (
                          <p className="text-xs text-green-600">
                            Paid: {formatCurrency(invoice.paid_amount)}
                          </p>
                        )}
                        {invoice.total_amount - invoice.paid_amount > 0 && (
                          <p className="text-xs text-red-600">
                            Due: {formatCurrency(invoice.total_amount - invoice.paid_amount)}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <Link
                          href={`/accounting/purchase-invoices/${invoice.id}`}
                          className="p-1.5 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          title="View Details"
                        >
                          <Eye className="w-4 h-4" />
                        </Link>
                        {invoice.status === 'draft' && (
                          <>
                            <Link
                              href={`/accounting/purchase-invoices/${invoice.id}/edit`}
                              className="p-1.5 text-gray-600 hover:text-emerald-600 hover:bg-emerald-50 rounded transition-colors"
                              title="Edit Invoice"
                            >
                              <Edit className="w-4 h-4" />
                            </Link>
                            <button
                              onClick={() => handleDeleteInvoice(invoice.id)}
                              className="p-1.5 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                              title="Delete Invoice"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </>
                        )}
                        {invoice.status === 'submitted' && (
                          <button
                            onClick={() => handleApproveInvoice(invoice.id)}
                            className="p-1.5 text-gray-600 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
                            title="Approve Invoice"
                          >
                            <CheckCircle className="w-4 h-4" />
                          </button>
                        )}
                        {(invoice.status === 'approved' || invoice.payment_status === 'partial') && (
                          <Link
                            href={`/accounting/purchase-invoices/${invoice.id}/payment`}
                            className="p-1.5 text-gray-600 hover:text-purple-600 hover:bg-purple-50 rounded transition-colors"
                            title="Record Payment"
                          >
                            <CreditCard className="w-4 h-4" />
                          </Link>
                        )}
                        {invoice.status !== 'paid' && invoice.status !== 'cancelled' && (
                          <button
                            onClick={() => handleCancelInvoice(invoice.id)}
                            className="p-1.5 text-gray-600 hover:text-orange-600 hover:bg-orange-50 rounded transition-colors"
                            title="Cancel Invoice"
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

        {filteredInvoices.length === 0 && (
          <div className="text-center py-12">
            <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">No purchase invoices found</p>
            <p className="text-sm text-gray-500 mt-2">
              {searchTerm || statusFilter !== 'all' || paymentStatusFilter !== 'all'
                ? 'Try adjusting your filters'
                : 'Create your first purchase invoice to get started'
              }
            </p>
            <Link
              href="/accounting/purchase-invoices/new"
              className="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-4 h-4" />
              <span>Create Purchase Invoice</span>
            </Link>
          </div>
        )}
      </div>

      {/* Overdue Alert */}
      {stats.overdue > 0 && (
        <div className="mt-6 bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-start gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-medium text-red-900">Overdue Invoices Alert</p>
              <p className="text-sm text-red-700 mt-1">
                You have {stats.overdue} overdue invoice{stats.overdue > 1 ? 's' : ''} totaling{' '}
                {formatCurrency(
                  invoices
                    .filter(i => i.payment_status === 'overdue')
                    .reduce((sum, i) => sum + (i.total_amount - i.paid_amount), 0)
                )}
                . Please review and process payments.
              </p>
            </div>
            <button
              onClick={() => setPaymentStatusFilter('overdue')}
              className="text-sm font-medium text-red-600 hover:text-red-700"
            >
              View Overdue
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
