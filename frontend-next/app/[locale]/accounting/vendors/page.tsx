'use client';

import { useEffect, useState } from 'react';
import { 
  Plus, Search, Filter, Download, Upload, Eye, Edit, Trash2,
  Building, Mail, Phone, MapPin, CreditCard, DollarSign,
  TrendingUp, TrendingDown, RefreshCw, AlertCircle, User,
  FileText, CheckCircle, XCircle
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface Vendor {
  id: string;
  vendor_code: string;
  vendor_name: string;
  vendor_type: 'supplier' | 'contractor' | 'service_provider';
  contact_person: string;
  email: string;
  phone: string;
  address: string;
  tax_id: string;
  payment_terms: number;
  credit_limit: number;
  current_balance: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
}

export default function VendorsPage() {
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [filteredVendors, setFilteredVendors] = useState<Vendor[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('active');
  const [selectedVendor, setSelectedVendor] = useState<Vendor | null>(null);
  const [showModal, setShowModal] = useState(false);

  useEffect(() => {
    fetchVendors();
  }, []);

  useEffect(() => {
    filterVendors();
  }, [vendors, searchTerm, typeFilter, statusFilter]);

  const fetchVendors = async () => {
    try {
      setLoading(true);
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/vendors`
      );
      
      if (!response.ok) {
        throw new Error('Failed to fetch vendors');
      }

      const data = await response.json();
      setVendors(data || []);
    } catch (error) {
      console.error('Failed to fetch vendors:', error);
      toast.error('Failed to load vendors');
    } finally {
      setLoading(false);
    }
  };

  const filterVendors = () => {
    let filtered = [...vendors];

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(vendor =>
        vendor.vendor_code.toLowerCase().includes(searchTerm.toLowerCase()) ||
        vendor.vendor_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        vendor.contact_person.toLowerCase().includes(searchTerm.toLowerCase()) ||
        vendor.email.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter(vendor => vendor.vendor_type === typeFilter);
    }

    // Status filter
    if (statusFilter === 'active') {
      filtered = filtered.filter(vendor => vendor.is_active);
    } else if (statusFilter === 'inactive') {
      filtered = filtered.filter(vendor => !vendor.is_active);
    }

    setFilteredVendors(filtered);
  };

  const handleDeleteVendor = async (vendorId: string) => {
    if (!confirm('Are you sure you want to deactivate this vendor?')) {
      return;
    }

    try {
      const response = await authenticatedFetch(
        `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/vendors/${vendorId}`,
        { method: 'DELETE' }
      );

      if (!response.ok) {
        throw new Error('Failed to delete vendor');
      }

      toast.success('Vendor deactivated successfully');
      fetchVendors();
    } catch (error) {
      console.error('Failed to delete vendor:', error);
      toast.error('Failed to deactivate vendor');
    }
  };

  const getTypeBadge = (type: string) => {
    const badges: Record<string, { bg: string; text: string }> = {
      supplier: { bg: 'bg-blue-100', text: 'text-blue-800' },
      contractor: { bg: 'bg-purple-100', text: 'text-purple-800' },
      service_provider: { bg: 'bg-emerald-100', text: 'text-emerald-800' }
    };
    const badge = badges[type] || { bg: 'bg-gray-100', text: 'text-gray-800' };
    
    return (
      <span className={`px-2.5 py-1 text-xs font-semibold rounded-full ${badge.bg} ${badge.text}`}>
        {type.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')}
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

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-emerald-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading vendors...</p>
        </div>
      </div>
    );
  }

  const stats = {
    total: vendors.length,
    active: vendors.filter(v => v.is_active).length,
    inactive: vendors.filter(v => !v.is_active).length,
    suppliers: vendors.filter(v => v.vendor_type === 'supplier').length,
    contractors: vendors.filter(v => v.vendor_type === 'contractor').length,
    serviceProviders: vendors.filter(v => v.vendor_type === 'service_provider').length,
    totalBalance: vendors.reduce((sum, v) => sum + v.current_balance, 0),
    totalCreditLimit: vendors.reduce((sum, v) => sum + v.credit_limit, 0)
  };

  return (
    <div className="p-8">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Vendors</h1>
            <p className="text-gray-600 mt-2">Manage your suppliers and service providers</p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/accounting/purchase-invoices"
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <FileText className="w-4 h-4" />
              <span>Invoices</span>
            </Link>
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
            <Link
              href="/accounting/vendors/new"
              className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
            >
              <Plus className="w-5 h-5" />
              <span>New Vendor</span>
            </Link>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Total Vendors</p>
          <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
        </div>
        <div className="bg-green-50 rounded-lg shadow-sm p-4 border border-green-200">
          <p className="text-xs text-green-700 mb-1">Active</p>
          <p className="text-2xl font-bold text-green-900">{stats.active}</p>
        </div>
        <div className="bg-gray-50 rounded-lg shadow-sm p-4 border border-gray-200">
          <p className="text-xs text-gray-600 mb-1">Inactive</p>
          <p className="text-2xl font-bold text-gray-400">{stats.inactive}</p>
        </div>
        <div className="bg-blue-50 rounded-lg shadow-sm p-4 border border-blue-200">
          <p className="text-xs text-blue-700 mb-1">Suppliers</p>
          <p className="text-2xl font-bold text-blue-900">{stats.suppliers}</p>
        </div>
        <div className="bg-purple-50 rounded-lg shadow-sm p-4 border border-purple-200">
          <p className="text-xs text-purple-700 mb-1">Contractors</p>
          <p className="text-2xl font-bold text-purple-900">{stats.contractors}</p>
        </div>
        <div className="bg-emerald-50 rounded-lg shadow-sm p-4 border border-emerald-200">
          <p className="text-xs text-emerald-700 mb-1">Service Providers</p>
          <p className="text-2xl font-bold text-emerald-900">{stats.serviceProviders}</p>
        </div>
        <div className="bg-orange-50 rounded-lg shadow-sm p-4 border border-orange-200">
          <p className="text-xs text-orange-700 mb-1">Total Balance</p>
          <p className="text-lg font-bold text-orange-900">{formatCurrency(stats.totalBalance)}</p>
        </div>
        <div className="bg-teal-50 rounded-lg shadow-sm p-4 border border-teal-200">
          <p className="text-xs text-teal-700 mb-1">Credit Limit</p>
          <p className="text-lg font-bold text-teal-900">{formatCurrency(stats.totalCreditLimit)}</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow-sm p-6 border border-gray-200 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search vendors..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              />
            </div>
          </div>
          <div>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Types</option>
              <option value="supplier">Supplier</option>
              <option value="contractor">Contractor</option>
              <option value="service_provider">Service Provider</option>
            </select>
          </div>
          <div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="all">All Status</option>
              <option value="active">Active</option>
              <option value="inactive">Inactive</option>
            </select>
          </div>
        </div>
      </div>

      {/* Vendors Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredVendors.map((vendor) => (
          <div
            key={vendor.id}
            className="bg-white rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow overflow-hidden"
          >
            {/* Card Header */}
            <div className="bg-gradient-to-r from-emerald-50 to-teal-50 px-6 py-4 border-b border-gray-200">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-1">
                    <Building className="w-5 h-5 text-emerald-600" />
                    <h3 className="text-lg font-bold text-gray-900 truncate">{vendor.vendor_name}</h3>
                  </div>
                  <p className="text-sm font-mono text-gray-600">{vendor.vendor_code}</p>
                </div>
                <div className="flex flex-col items-end gap-2">
                  {getTypeBadge(vendor.vendor_type)}
                  {vendor.is_active ? (
                    <span className="flex items-center gap-1 text-xs text-green-600">
                      <CheckCircle className="w-3 h-3" />
                      Active
                    </span>
                  ) : (
                    <span className="flex items-center gap-1 text-xs text-red-600">
                      <XCircle className="w-3 h-3" />
                      Inactive
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Card Body */}
            <div className="p-6 space-y-3">
              {vendor.contact_person && (
                <div className="flex items-center gap-2 text-sm text-gray-700">
                  <Building className="w-4 h-4 text-gray-400" />
                  <span>{vendor.contact_person}</span>
                </div>
              )}
              {vendor.email && (
                <div className="flex items-center gap-2 text-sm text-gray-700">
                  <Mail className="w-4 h-4 text-gray-400" />
                  <span className="truncate">{vendor.email}</span>
                </div>
              )}
              {vendor.phone && (
                <div className="flex items-center gap-2 text-sm text-gray-700">
                  <Phone className="w-4 h-4 text-gray-400" />
                  <span>{vendor.phone}</span>
                </div>
              )}
              {vendor.address && (
                <div className="flex items-start gap-2 text-sm text-gray-700">
                  <MapPin className="w-4 h-4 text-gray-400 mt-0.5" />
                  <span className="line-clamp-2">{vendor.address}</span>
                </div>
              )}

              {/* Financial Info */}
              <div className="pt-3 border-t border-gray-200 space-y-2">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">Payment Terms</span>
                  <span className="font-medium text-gray-900">{vendor.payment_terms} days</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">Credit Limit</span>
                  <span className="font-medium text-gray-900">{formatCurrency(vendor.credit_limit)}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600">Current Balance</span>
                  <span className={`font-bold ${vendor.current_balance > 0 ? 'text-orange-600' : 'text-gray-900'}`}>
                    {formatCurrency(vendor.current_balance)}
                  </span>
                </div>
              </div>
            </div>

            {/* Card Footer */}
            <div className="bg-gray-50 px-6 py-3 border-t border-gray-200 flex items-center justify-between">
              <Link
                href={`/accounting/vendors/${vendor.id}`}
                className="flex items-center gap-1 text-sm text-emerald-600 hover:text-emerald-700 font-medium"
              >
                <Eye className="w-4 h-4" />
                View Details
              </Link>
              <div className="flex items-center gap-2">
                <Link
                  href={`/accounting/vendors/${vendor.id}/edit`}
                  className="p-1.5 text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded transition-colors"
                  title="Edit Vendor"
                >
                  <Edit className="w-4 h-4" />
                </Link>
                <button
                  onClick={() => handleDeleteVendor(vendor.id)}
                  className="p-1.5 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded transition-colors"
                  title="Deactivate Vendor"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {filteredVendors.length === 0 && (
        <div className="text-center py-12 bg-white rounded-lg shadow-sm border border-gray-200">
          <Building className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-600">No vendors found</p>
          <p className="text-sm text-gray-500 mt-2">
            {searchTerm || typeFilter !== 'all' || statusFilter !== 'all'
              ? 'Try adjusting your filters'
              : 'Create your first vendor to get started'
            }
          </p>
          <Link
            href="/accounting/vendors/new"
            className="inline-flex items-center gap-2 mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"
          >
            <Plus className="w-4 h-4" />
            <span>Create Vendor</span>
          </Link>
        </div>
      )}
    </div>
  );
}
