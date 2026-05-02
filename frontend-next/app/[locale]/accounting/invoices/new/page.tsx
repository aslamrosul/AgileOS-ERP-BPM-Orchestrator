'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Save, X, Plus, Trash2, Building, Calendar, FileText, DollarSign, AlertCircle, ArrowLeft, Calculator } from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';
import Link from 'next/link';

interface Customer { id: string; customer_code: string; customer_name: string; customer_type: string; payment_terms: number; current_balance: number; }
interface Account { id: string; account_code: string; account_name: string; account_type: string; }
interface InvoiceLine { id: string; description: string; account_id: string; quantity: number; unit_price: number; tax_rate: number; amount: number; }

export default function NewSalesInvoicePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [customerId, setCustomerId] = useState('');
  const [invoiceDate, setInvoiceDate] = useState(new Date().toISOString().split('T')[0]);
  const [dueDate, setDueDate] = useState('');
  const [reference, setReference] = useState('');
  const [description, setDescription] = useState('');
  const [lines, setLines] = useState<InvoiceLine[]>([{ id: crypto.randomUUID(), description: '', account_id: '', quantity: 1, unit_price: 0, tax_rate: 11, amount: 0 }]);

  useEffect(() => { fetchData(); }, []);
  useEffect(() => {
    if (customerId && invoiceDate) {
      const customer = customers.find(c => c.id === customerId);
      if (customer && customer.payment_terms) {
        const date = new Date(invoiceDate);
        date.setDate(date.getDate() + customer.payment_terms);
        setDueDate(date.toISOString().split('T')[0]);
      }
    }
  }, [customerId, invoiceDate, customers]);

  const fetchData = async () => {
    try {
      const [customersRes, accountsRes] = await Promise.all([
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/customers`),
        authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/accounts`)
      ]);
      if (customersRes.ok && accountsRes.ok) {
        const [customersData, accountsData] = await Promise.all([customersRes.json(), accountsRes.json()]);
        setCustomers(customersData || []);
        setAccounts((accountsData || []).filter((a: Account) => a.account_type === 'revenue'));
      }
    } catch (error) { toast.error('Failed to load form data'); }
  };

  const addLine = () => { setLines([...lines, { id: crypto.randomUUID(), description: '', account_id: '', quantity: 1, unit_price: 0, tax_rate: 11, amount: 0 }]); };
  const removeLine = (id: string) => { if (lines.length > 1) setLines(lines.filter(line => line.id !== id)); };
  const updateLine = (id: string, field: keyof InvoiceLine, value: any) => {
    setLines(lines.map(line => {
      if (line.id === id) {
        const updated = { ...line, [field]: value };
        if (field === 'quantity' || field === 'unit_price' || field === 'tax_rate') {
          const subtotal = updated.quantity * updated.unit_price;
          const tax = subtotal * (updated.tax_rate / 100);
          updated.amount = subtotal + tax;
        }
        return updated;
      }
      return line;
    }));
  };

  const calculateTotals = () => {
    const subtotal = lines.reduce((sum, line) => sum + (line.quantity * line.unit_price), 0);
    const taxAmount = lines.reduce((sum, line) => sum + ((line.quantity * line.unit_price) * (line.tax_rate / 100)), 0);
    return { subtotal, taxAmount, total: subtotal + taxAmount };
  };

  const handleSubmit = async (status: 'draft' | 'submitted') => {
    if (!customerId) { toast.error('Please select a customer'); return; }
    if (!invoiceDate || !dueDate) { toast.error('Please enter invoice and due dates'); return; }
    if (lines.some(line => !line.description || !line.account_id)) { toast.error('Please fill in all line items'); return; }
    const totals = calculateTotals();
    const customer = customers.find(c => c.id === customerId);
    const payload = {
      customer_id: customerId, customer_name: customer?.customer_name || '', invoice_date: new Date(invoiceDate).toISOString(),
      due_date: new Date(dueDate).toISOString(), reference: reference || undefined, description: description || undefined,
      status, total_amount: totals.total, tax_amount: totals.taxAmount, discount_amount: 0,
      lines: lines.map(line => ({ description: line.description, account_id: line.account_id, quantity: line.quantity, unit_price: line.unit_price, tax_rate: line.tax_rate, amount: line.amount }))
    };
    try {
      setLoading(true);
      const response = await authenticatedFetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting/sales-invoices`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload)
      });
      if (!response.ok) throw new Error('Failed to create invoice');
      const data = await response.json();
      toast.success(`Sales invoice ${status === 'draft' ? 'saved as draft' : 'submitted'} successfully`);
      router.push(`/accounting/invoices`);
    } catch (error) { toast.error('Failed to create sales invoice'); } finally { setLoading(false); }
  };

  const formatCurrency = (amount: number) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(amount);
  const totals = calculateTotals();
  const selectedCustomer = customers.find(c => c.id === customerId);

  return (
    <div className="p-8">
      <div className="mb-8">
        <div className="flex items-center gap-4 mb-4">
          <Link href="/accounting/invoices" className="p-2 hover:bg-gray-100 rounded-lg transition-colors"><ArrowLeft className="w-5 h-5" /></Link>
          <div><h1 className="text-3xl font-bold text-gray-900">New Sales Invoice</h1><p className="text-gray-600 mt-1">Create a new customer invoice</p></div>
        </div>
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2"><Building className="w-5 h-5 text-emerald-600" />Customer & Date Information</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div><label className="block text-sm font-medium text-gray-700 mb-1">Customer <span className="text-red-500">*</span></label>
                <select value={customerId} onChange={(e) => setCustomerId(e.target.value)} className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required>
                  <option value="">Select Customer</option>
                  {customers.map(customer => (<option key={customer.id} value={customer.id}>{customer.customer_name} ({customer.customer_code})</option>))}
                </select>
                {selectedCustomer && <p className="text-xs text-gray-500 mt-1">Payment Terms: {selectedCustomer.payment_terms} days</p>}
              </div>
              <div><label className="block text-sm font-medium text-gray-700 mb-1">Reference Number</label>
                <input type="text" value={reference} onChange={(e) => setReference(e.target.value)} placeholder="Customer PO number" className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" />
              </div>
              <div><label className="block text-sm font-medium text-gray-700 mb-1">Invoice Date <span className="text-red-500">*</span></label>
                <input type="date" value={invoiceDate} onChange={(e) => setInvoiceDate(e.target.value)} className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required />
              </div>
              <div><label className="block text-sm font-medium text-gray-700 mb-1">Due Date <span className="text-red-500">*</span></label>
                <input type="date" value={dueDate} onChange={(e) => setDueDate(e.target.value)} className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required />
              </div>
              <div className="md:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                <textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Invoice description or notes" rows={2} className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" />
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2"><FileText className="w-5 h-5 text-emerald-600" />Invoice Lines</h2>
              <button onClick={addLine} className="flex items-center gap-2 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors"><Plus className="w-4 h-4" />Add Line</button>
            </div>
            <div className="space-y-4">
              {lines.map((line, index) => (
                <div key={line.id} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex items-start justify-between mb-3">
                    <span className="text-sm font-medium text-gray-700">Line {index + 1}</span>
                    {lines.length > 1 && <button onClick={() => removeLine(line.id)} className="p-1 text-red-600 hover:bg-red-50 rounded transition-colors"><Trash2 className="w-4 h-4" /></button>}
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <div className="md:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Description <span className="text-red-500">*</span></label>
                      <input type="text" value={line.description} onChange={(e) => updateLine(line.id, 'description', e.target.value)} placeholder="Item description" className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required />
                    </div>
                    <div className="md:col-span-2"><label className="block text-sm font-medium text-gray-700 mb-1">Revenue Account <span className="text-red-500">*</span></label>
                      <select value={line.account_id} onChange={(e) => updateLine(line.id, 'account_id', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" required>
                        <option value="">Select Account</option>
                        {accounts.map(account => (<option key={account.id} value={account.id}>{account.account_code} - {account.account_name}</option>))}
                      </select>
                    </div>
                    <div><label className="block text-sm font-medium text-gray-700 mb-1">Quantity</label>
                      <input type="number" value={line.quantity} onChange={(e) => updateLine(line.id, 'quantity', parseFloat(e.target.value) || 0)} min="0" step="0.01" className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" />
                    </div>
                    <div><label className="block text-sm font-medium text-gray-700 mb-1">Unit Price (IDR)</label>
                      <input type="number" value={line.unit_price} onChange={(e) => updateLine(line.id, 'unit_price', parseFloat(e.target.value) || 0)} min="0" step="0.01" className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" />
                    </div>
                    <div><label className="block text-sm font-medium text-gray-700 mb-1">Tax Rate (%)</label>
                      <input type="number" value={line.tax_rate} onChange={(e) => updateLine(line.id, 'tax_rate', parseFloat(e.target.value) || 0)} min="0" max="100" step="0.01" className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500" />
                    </div>
                    <div><label className="block text-sm font-medium text-gray-700 mb-1">Line Total</label>
                      <div className="px-3 py-2 bg-gray-50 border border-gray-300 rounded-lg text-gray-900 font-medium">{formatCurrency(line.amount)}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 sticky top-8">
            <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2"><Calculator className="w-5 h-5 text-emerald-600" />Invoice Summary</h2>
            <div className="space-y-3 mb-6">
              <div className="flex justify-between text-sm"><span className="text-gray-600">Subtotal:</span><span className="font-medium text-gray-900">{formatCurrency(totals.subtotal)}</span></div>
              <div className="flex justify-between text-sm"><span className="text-gray-600">Tax Amount:</span><span className="font-medium text-gray-900">{formatCurrency(totals.taxAmount)}</span></div>
              <div className="border-t border-gray-200 pt-3"><div className="flex justify-between"><span className="text-base font-semibold text-gray-900">Total Amount:</span><span className="text-lg font-bold text-emerald-600">{formatCurrency(totals.total)}</span></div></div>
            </div>
            {selectedCustomer && (
              <div className="mb-6 p-3 bg-blue-50 border border-blue-200 rounded-lg"><p className="text-xs font-medium text-blue-900 mb-1">Customer Balance</p><p className="text-sm font-semibold text-blue-700">{formatCurrency(selectedCustomer.current_balance)}</p></div>
            )}
            <div className="space-y-3">
              <button onClick={() => handleSubmit('submitted')} disabled={loading} className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                <Save className="w-4 h-4" />{loading ? 'Saving...' : 'Submit Invoice'}
              </button>
              <button onClick={() => handleSubmit('draft')} disabled={loading} className="w-full flex items-center justify-center gap-2 px-4 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                <FileText className="w-4 h-4" />Save as Draft
              </button>
              <Link href="/accounting/invoices" className="w-full flex items-center justify-center gap-2 px-4 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
                <X className="w-4 h-4" />Cancel
              </Link>
            </div>
            <div className="mt-6 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
              <div className="flex items-start gap-2"><AlertCircle className="w-4 h-4 text-yellow-600 flex-shrink-0 mt-0.5" />
                <div><p className="text-xs font-medium text-yellow-900">Note</p><p className="text-xs text-yellow-700 mt-1">Draft invoices can be edited later. Submitted invoices require approval before payment.</p></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
