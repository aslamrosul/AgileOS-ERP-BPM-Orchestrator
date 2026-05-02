'use client';
import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Save, X, Building, User, CreditCard } from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

export default function NewVendorPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    vendor_name: '', vendor_type: 'supplier', tax_id: '', contact_person: '',
    email: '', phone: '', address: '', payment_terms: 30, credit_limit: 0, is_active: true
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.vendor_name.trim()) { toast.error('Vendor name is required'); return; }
    try {
      setLoading(true);
      const response = await authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/vendors`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(formData)
      });
      if (!response.ok) throw new Error('Failed to create vendor');
      toast.success('Vendor created successfully');
      router.push('/accounting/vendors');
    } catch (error: any) {
      toast.error(error.message || 'Failed to create vendor');
    } finally { setLoading(false); }
  };

  return (
    <div className="p-8">
      <div className="mb-8">
        <Link href="/accounting/vendors" className="inline-flex items-center gap-2 text-emerald-600 hover:text-emerald-700 mb-4">
          <ArrowLeft className="w-4 h-4" /><span>Back to Vendors</span>
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">New Vendor</h1>
        <p className="text-gray-600 mt-2">Create a new vendor/supplier</p>
      </div>
      <form onSubmit={handleSubmit} className="max-w-4xl space-y-6">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Building className="w-5 h-5 text-emerald-600" />Basic Information
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Vendor Name <span className="text-red-500">*</span></label>
              <input type="text" value={formData.vendor_name} onChange={(e) => setFormData({...formData, vendor_name: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Vendor Type</label>
              <select value={formData.vendor_type} onChange={(e) => setFormData({...formData, vendor_type: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500">
                <option value="supplier">Supplier</option><option value="contractor">Contractor</option>
                <option value="service_provider">Service Provider</option><option value="other">Other</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tax ID / NPWP</label>
              <input type="text" value={formData.tax_id} onChange={(e) => setFormData({...formData, tax_id: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" />
            </div>
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <User className="w-5 h-5 text-emerald-600" />Contact Information
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><label className="block text-sm font-medium text-gray-700 mb-1">Contact Person</label>
              <input type="text" value={formData.contact_person} onChange={(e) => setFormData({...formData, contact_person: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" /></div>
            <div><label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
              <input type="email" value={formData.email} onChange={(e) => setFormData({...formData, email: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" /></div>
            <div><label className="block text-sm font-medium text-gray-700 mb-1">Phone</label>
              <input type="tel" value={formData.phone} onChange={(e) => setFormData({...formData, phone: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" /></div>
            <div><label className="block text-sm font-medium text-gray-700 mb-1">Address</label>
              <input type="text" value={formData.address} onChange={(e) => setFormData({...formData, address: e.target.value})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" /></div>
          </div>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <CreditCard className="w-5 h-5 text-emerald-600" />Financial Information
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div><label className="block text-sm font-medium text-gray-700 mb-1">Payment Terms (days)</label>
              <input type="number" value={formData.payment_terms} onChange={(e) => setFormData({...formData, payment_terms: parseInt(e.target.value) || 0})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" min="0" /></div>
            <div><label className="block text-sm font-medium text-gray-700 mb-1">Credit Limit (IDR)</label>
              <input type="number" value={formData.credit_limit} onChange={(e) => setFormData({...formData, credit_limit: parseFloat(e.target.value) || 0})}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" min="0" step="1000" /></div>
            <div className="md:col-span-2">
              <label className="flex items-center gap-2">
                <input type="checkbox" checked={formData.is_active} onChange={(e) => setFormData({...formData, is_active: e.target.checked})}
                  className="w-4 h-4 text-emerald-600 border-gray-300 rounded focus:ring-emerald-500" />
                <span className="text-sm font-medium text-gray-700">Active Vendor</span>
              </label>
            </div>
          </div>
        </div>
        <div className="flex gap-3">
          <Link href="/accounting/vendors" className="flex-1 flex items-center justify-center gap-2 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
            <X className="w-4 h-4" />Cancel
          </Link>
          <button type="submit" disabled={loading}
            className="flex-1 flex items-center justify-center gap-2 px-6 py-3 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors disabled:opacity-50">
            {loading ? <><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Creating...</> : <><Save className="w-4 h-4" />Create Vendor</>}
          </button>
        </div>
      </form>
    </div>
  );
}
