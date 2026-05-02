'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Save, X, DollarSign, Calendar, FileText, CreditCard, Building, User } from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface Vendor { id: string; vendor_code: string; vendor_name: string; }
interface Customer { id: string; customer_code: string; customer_name: string; }

export default function NewPaymentPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [formData, setFormData] = useState({
    payment_type: 'vendor_payment' as 'vendor_payment' | 'customer_receipt',
    party_id: '',
    party_name: '',
    payment_method: 'bank_transfer' as 'cash' | 'bank_transfer' | 'check' | 'credit_card',
    amount: 0,
    payment_date: new Date().toISOString().split('T')[0],
    reference: '',
    description: ''
  });

  useEffect(() => { fetchData(); }, []);

  const fetchData = async () => {
    try {
      const [vendorsRes, customersRes] = await Promise.all([
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/vendors`),
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/customers`)
      ]);
      if (vendorsRes.ok && customersRes.ok) {
        const [vendorsData, customersData] = await Promise.all([vendorsRes.json(), customersRes.json()]);
        setVendors(vendorsData || []);
        setCustomers(customersData || []);
      }
    } catch (error) { toast.error('Failed to load data'); }
  };

  const handlePartyChange = (partyId: string) => {
    const parties = formData.payment_type === 'vendor_payment' ? vendors : customers;
    const party = parties.find(p => p.id === partyId);
    setFormData({
      ...formData,
      party_id: partyId,
      party_name: party ? ('vendor_name' in party ? party.vendor_name : party.customer_name) : ''
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.party_id) { toast.error('Please select a party'); return; }
    if (formData.amount <= 0) { toast.error('Amount must be greater than 0'); return; }
    
    try {
      setLoading(true);
      const response = await authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/payments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...formData,
          payment_date: new Date(formData.payment_date).toISOString()
        })
      });
      if (!response.ok) throw new Error('Failed to create payment');
      toast.success('Payment created successfully');
      router.push('/accounting/payments');
    } catch (error: any) {
      toast.error(error.message || 'Failed to create payment');
    } finally { setLoading(false); }
  };

  const formatCurrency = (amount: number) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(amount);
  const parties = formData.payment_type === 'vendor_payment' ? vendors : customers;

  return (
    <div className="p-8">
      <div className="mb-8">
        <Link href="/accounting/payments" className="inline-flex items-center gap-2 text-emerald-600 hover:text-emerald-700 mb-4">
          <ArrowLeft className="w-4 h-4" /><span>Back to Payments</span>
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">New Payment</h1>
        <p className="text-gray-600 mt-2">Record a new payment transaction</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <DollarSign className="w-5 h-5 text-emerald-600" />Payment Type
              </h2>
              <div className="grid grid-cols-2 gap-4">
                <label className={`flex items-center gap-3 p-4 border-2 rounded-lg cursor-pointer transition-all ${formData.payment_type === 'vendor_payment' ? 'border-emerald-600 bg-emerald-50' : 'border-gray-200 hover:border-gray-300'}`}>
                  <input type="radio" name="payment_type" value="vendor_payment" checked={formData.payment_type === 'vendor_payment'}
                    onChange={(e) => setFormData({...formData, payment_type: e.target.value as any, party_id: '', party_name: ''})}
                    className="w-4 h-4 text-emerald-600" />
                  <div><p className="font-medium text-gray-900">Vendor Payment</p><p className="text-xs text-gray-600">Pay to vendor</p></div>
                </label>
                <label className={`flex items-center gap-3 p-4 border-2 rounded-lg cursor-pointer transition-all ${formData.payment_type === 'customer_receipt' ? 'border-emerald-600 bg-emerald-50' : 'border-gray-200 hover:border-gray-300'}`}>
                  <input type="radio" name="payment_type" value="customer_receipt" checked={formData.payment_type === 'customer_receipt'}
                    onChange={(e) => setFormData({...formData, payment_type: e.target.value as any, party_id: '', party_name: ''})}
                    className="w-4 h-4 text-emerald-600" />
                  <div><p className="font-medium text-gray-900">Customer Receipt</p><p className="text-xs text-gray-600">Receive from customer</p></div>
                </label>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                {formData.payment_type === 'vendor_payment' ? <Building className="w-5 h-5 text-emerald-600" /> : <User className="w-5 h-5 text-emerald-600" />}
                {formData.payment_type === 'vendor_payment' ? 'Vendor' : 'Customer'} Information
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    {formData.payment_type === 'vendor_payment' ? 'Vendor' : 'Customer'} <span className="text-red-500">*</span>
                  </label>
                  <select value={formData.party_id} onChange={(e) => handlePartyChange(e.target.value)}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required>
                    <option value="">Select {formData.payment_type === 'vendor_payment' ? 'Vendor' : 'Customer'}</option>
                    {parties.map(party => (
                      <option key={party.id} value={party.id}>
                        {'vendor_name' in party ? `${party.vendor_name} (${party.vendor_code})` : `${party.customer_name} (${party.customer_code})`}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <CreditCard className="w-5 h-5 text-emerald-600" />Payment Details
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Payment Method <span className="text-red-500">*</span></label>
                  <select value={formData.payment_method} onChange={(e) => setFormData({...formData, payment_method: e.target.value as any})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required>
                    <option value="cash">Cash</option>
                    <option value="bank_transfer">Bank Transfer</option>
                    <option value="check">Check</option>
                    <option value="credit_card">Credit Card</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Payment Date <span className="text-red-500">*</span></label>
                  <input type="date" value={formData.payment_date} onChange={(e) => setFormData({...formData, payment_date: e.target.value})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Amount (IDR) <span className="text-red-500">*</span></label>
                  <input type="number" value={formData.amount} onChange={(e) => setFormData({...formData, amount: parseFloat(e.target.value) || 0})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" min="0" step="1000" required />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Reference Number</label>
                  <input type="text" value={formData.reference} onChange={(e) => setFormData({...formData, reference: e.target.value})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" placeholder="e.g., TRX-2026-001" />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                  <textarea value={formData.description} onChange={(e) => setFormData({...formData, description: e.target.value})}
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" rows={3} placeholder="Payment notes or description" />
                </div>
              </div>
            </div>

            <div className="flex gap-3">
              <Link href="/accounting/payments" className="flex-1 flex items-center justify-center gap-2 px-6 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
                <X className="w-4 h-4" />Cancel
              </Link>
              <button type="submit" disabled={loading}
                className="flex-1 flex items-center justify-center gap-2 px-6 py-3 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors disabled:opacity-50">
                {loading ? <><div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>Creating...</> : <><Save className="w-4 h-4" />Create Payment</>}
              </button>
            </div>
          </form>
        </div>

        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 sticky top-8">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Payment Summary</h2>
            <div className="space-y-3">
              <div className="p-4 bg-gray-50 rounded-lg">
                <p className="text-sm text-gray-600 mb-1">Payment Type</p>
                <p className="font-semibold text-gray-900">{formData.payment_type === 'vendor_payment' ? 'Vendor Payment' : 'Customer Receipt'}</p>
              </div>
              <div className="p-4 bg-gray-50 rounded-lg">
                <p className="text-sm text-gray-600 mb-1">Payment Method</p>
                <p className="font-semibold text-gray-900 capitalize">{formData.payment_method.replace('_', ' ')}</p>
              </div>
              <div className="p-4 bg-emerald-50 rounded-lg border border-emerald-200">
                <p className="text-sm text-emerald-700 mb-1">Amount</p>
                <p className="text-2xl font-bold text-emerald-900">{formatCurrency(formData.amount)}</p>
              </div>
              {formData.party_name && (
                <div className="p-4 bg-blue-50 rounded-lg border border-blue-200">
                  <p className="text-sm text-blue-700 mb-1">{formData.payment_type === 'vendor_payment' ? 'Vendor' : 'Customer'}</p>
                  <p className="font-semibold text-blue-900">{formData.party_name}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
