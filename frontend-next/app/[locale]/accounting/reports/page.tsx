'use client';

import { useState } from 'react';
import { 
  PieChart, FileText, TrendingUp, Download, X, Calendar, DollarSign,
  BarChart3, Activity, Users, Building, AlertCircle, CheckCircle
} from 'lucide-react';
import { authenticatedFetch } from '@/lib/auth';
import { toast } from 'sonner';

interface ReportConfig {
  name: string;
  description: string;
  icon: any;
  color: string;
  endpoint: string;
  requiresDateRange: boolean;
  requiresSingleDate: boolean;
}

export default function ReportsPage() {
  const [selectedReport, setSelectedReport] = useState<ReportConfig | null>(null);
  const [loading, setLoading] = useState(false);
  const [reportData, setReportData] = useState<any>(null);
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [asOfDate, setAsOfDate] = useState(new Date().toISOString().split('T')[0]);

  const reports: ReportConfig[] = [
    {
      name: 'Balance Sheet',
      description: 'Assets, Liabilities, and Equity',
      icon: PieChart,
      color: 'blue',
      endpoint: '/reports/balance-sheet',
      requiresDateRange: false,
      requiresSingleDate: true
    },
    {
      name: 'Profit & Loss',
      description: 'Revenue and Expenses',
      icon: TrendingUp,
      color: 'green',
      endpoint: '/reports/profit-loss',
      requiresDateRange: true,
      requiresSingleDate: false
    },
    {
      name: 'Cash Flow Statement',
      description: 'Operating, Investing, Financing',
      icon: Activity,
      color: 'purple',
      endpoint: '/reports/cash-flow',
      requiresDateRange: true,
      requiresSingleDate: false
    },
    {
      name: 'Trial Balance',
      description: 'Debit and Credit Balances',
      icon: BarChart3,
      color: 'orange',
      endpoint: '/reports/trial-balance',
      requiresDateRange: false,
      requiresSingleDate: true
    },
    {
      name: 'General Ledger',
      description: 'Detailed Transaction History',
      icon: FileText,
      color: 'teal',
      endpoint: '/reports/general-ledger',
      requiresDateRange: true,
      requiresSingleDate: false
    },
    {
      name: 'Aging Report (AR)',
      description: 'Accounts Receivable Aging',
      icon: Users,
      color: 'red',
      endpoint: '/reports/ar-aging',
      requiresDateRange: false,
      requiresSingleDate: true
    },
    {
      name: 'Aging Report (AP)',
      description: 'Accounts Payable Aging',
      icon: Building,
      color: 'yellow',
      endpoint: '/reports/ap-aging',
      requiresDateRange: false,
      requiresSingleDate: true
    }
  ];

  const getColorClasses = (color: string) => {
    const colors: Record<string, string> = {
      blue: 'bg-blue-100 text-blue-600',
      green: 'bg-green-100 text-green-600',
      purple: 'bg-purple-100 text-purple-600',
      orange: 'bg-orange-100 text-orange-600',
      teal: 'bg-teal-100 text-teal-600',
      red: 'bg-red-100 text-red-600',
      yellow: 'bg-yellow-100 text-yellow-600',
      indigo: 'bg-indigo-100 text-indigo-600'
    };
    return colors[color] || 'bg-gray-100 text-gray-600';
  };

  const handleGenerateReport = async () => {
    if (!selectedReport) return;

    if (selectedReport.requiresDateRange && (!fromDate || !toDate)) {
      toast.error('Please select date range');
      return;
    }

    if (selectedReport.requiresSingleDate && !asOfDate) {
      toast.error('Please select date');
      return;
    }

    try {
      setLoading(true);
      
      let url = `${process.env.NEXT_PUBLIC_API_URL}/api/v1/accounting${selectedReport.endpoint}`;
      
      if (selectedReport.requiresDateRange) {
        url += `?from_date=${fromDate}&to_date=${toDate}`;
      } else if (selectedReport.requiresSingleDate) {
        url += `?as_of_date=${asOfDate}`;
      }

      const response = await authenticatedFetch(url);

      if (!response.ok) {
        throw new Error('Failed to generate report');
      }

      const data = await response.json();
      setReportData(data);
      toast.success('Report generated successfully');
    } catch (error) {
      console.error('Failed to generate report:', error);
      toast.error('Failed to generate report');
    } finally {
      setLoading(false);
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

  const closeModal = () => {
    setSelectedReport(null);
    setReportData(null);
    setFromDate('');
    setToDate('');
  };

  const renderReportContent = () => {
    if (!reportData) return null;

    switch (selectedReport?.name) {
      case 'Balance Sheet':
        return (
          <div className="space-y-6">
            <div className="text-center mb-6">
              <h3 className="text-xl font-bold">Balance Sheet</h3>
              <p className="text-gray-600">As of {formatDate(reportData.as_of_date)}</p>
            </div>

            {/* Assets */}
            <div>
              <h4 className="font-semibold text-lg mb-3 text-blue-600">Assets</h4>
              <table className="w-full text-sm">
                <tbody>
                  {reportData.assets?.accounts?.map((account: any, idx: number) => (
                    <tr key={idx} className="border-b">
                      <td className="py-2">{account.account_name}</td>
                      <td className="py-2 text-right">{formatCurrency(account.balance)}</td>
                    </tr>
                  ))}
                  <tr className="font-bold bg-blue-50">
                    <td className="py-2">Total Assets</td>
                    <td className="py-2 text-right">{formatCurrency(reportData.total_assets)}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            {/* Liabilities */}
            <div>
              <h4 className="font-semibold text-lg mb-3 text-red-600">Liabilities</h4>
              <table className="w-full text-sm">
                <tbody>
                  {reportData.liabilities?.accounts?.map((account: any, idx: number) => (
                    <tr key={idx} className="border-b">
                      <td className="py-2">{account.account_name}</td>
                      <td className="py-2 text-right">{formatCurrency(account.balance)}</td>
                    </tr>
                  ))}
                  <tr className="font-bold bg-red-50">
                    <td className="py-2">Total Liabilities</td>
                    <td className="py-2 text-right">{formatCurrency(reportData.total_liabilities)}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            {/* Equity */}
            <div>
              <h4 className="font-semibold text-lg mb-3 text-green-600">Equity</h4>
              <table className="w-full text-sm">
                <tbody>
                  {reportData.equity?.accounts?.map((account: any, idx: number) => (
                    <tr key={idx} className="border-b">
                      <td className="py-2">{account.account_name}</td>
                      <td className="py-2 text-right">{formatCurrency(account.balance)}</td>
                    </tr>
                  ))}
                  <tr className="font-bold bg-green-50">
                    <td className="py-2">Total Equity</td>
                    <td className="py-2 text-right">{formatCurrency(reportData.total_equity)}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        );

      case 'Profit & Loss':
        return (
          <div className="space-y-6">
            <div className="text-center mb-6">
              <h3 className="text-xl font-bold">Profit & Loss Statement</h3>
              <p className="text-gray-600">
                {formatDate(reportData.from_date)} - {formatDate(reportData.to_date)}
              </p>
            </div>

            {/* Revenue */}
            <div>
              <h4 className="font-semibold text-lg mb-3 text-green-600">Revenue</h4>
              <table className="w-full text-sm">
                <tbody>
                  {reportData.revenue?.map((account: any, idx: number) => (
                    <tr key={idx} className="border-b">
                      <td className="py-2">{account.account_name}</td>
                      <td className="py-2 text-right">{formatCurrency(account.balance)}</td>
                    </tr>
                  ))}
                  <tr className="font-bold bg-green-50">
                    <td className="py-2">Total Revenue</td>
                    <td className="py-2 text-right">{formatCurrency(reportData.total_revenue)}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            {/* Expenses */}
            <div>
              <h4 className="font-semibold text-lg mb-3 text-red-600">Expenses</h4>
              <table className="w-full text-sm">
                <tbody>
                  {reportData.expenses?.map((account: any, idx: number) => (
                    <tr key={idx} className="border-b">
                      <td className="py-2">{account.account_name}</td>
                      <td className="py-2 text-right">{formatCurrency(account.balance)}</td>
                    </tr>
                  ))}
                  <tr className="font-bold bg-red-50">
                    <td className="py-2">Total Expenses</td>
                    <td className="py-2 text-right">{formatCurrency(reportData.total_expenses)}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            {/* Net Profit */}
            <div className={`p-4 rounded-lg ${reportData.net_profit >= 0 ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex justify-between items-center">
                <span className="font-bold text-lg">Net Profit</span>
                <span className={`font-bold text-xl ${reportData.net_profit >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                  {formatCurrency(reportData.net_profit)}
                </span>
              </div>
            </div>
          </div>
        );

      case 'Trial Balance':
        return (
          <div className="space-y-6">
            <div className="text-center mb-6">
              <h3 className="text-xl font-bold">Trial Balance</h3>
              <p className="text-gray-600">As of {formatDate(reportData.as_of_date)}</p>
            </div>

            <table className="w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="py-2 px-4 text-left">Account Code</th>
                  <th className="py-2 px-4 text-left">Account Name</th>
                  <th className="py-2 px-4 text-right">Debit</th>
                  <th className="py-2 px-4 text-right">Credit</th>
                </tr>
              </thead>
              <tbody>
                {reportData.accounts?.map((account: any, idx: number) => (
                  <tr key={idx} className="border-b">
                    <td className="py-2 px-4">{account.account_code}</td>
                    <td className="py-2 px-4">{account.account_name}</td>
                    <td className="py-2 px-4 text-right">{account.debit > 0 ? formatCurrency(account.debit) : '-'}</td>
                    <td className="py-2 px-4 text-right">{account.credit > 0 ? formatCurrency(account.credit) : '-'}</td>
                  </tr>
                ))}
                <tr className="font-bold bg-gray-50">
                  <td colSpan={2} className="py-2 px-4">Total</td>
                  <td className="py-2 px-4 text-right">{formatCurrency(reportData.total_debit)}</td>
                  <td className="py-2 px-4 text-right">{formatCurrency(reportData.total_credit)}</td>
                </tr>
              </tbody>
            </table>

            <div className={`p-4 rounded-lg ${reportData.is_balanced ? 'bg-green-50' : 'bg-red-50'}`}>
              <div className="flex items-center gap-2">
                {reportData.is_balanced ? (
                  <>
                    <CheckCircle className="w-5 h-5 text-green-600" />
                    <span className="font-semibold text-green-600">Trial Balance is Balanced</span>
                  </>
                ) : (
                  <>
                    <AlertCircle className="w-5 h-5 text-red-600" />
                    <span className="font-semibold text-red-600">Trial Balance is NOT Balanced</span>
                  </>
                )}
              </div>
            </div>
          </div>
        );

      case 'Aging Report (AR)':
      case 'Aging Report (AP)':
        return (
          <div className="space-y-6">
            <div className="text-center mb-6">
              <h3 className="text-xl font-bold">{selectedReport.name}</h3>
              <p className="text-gray-600">As of {formatDate(reportData.as_of_date)}</p>
            </div>

            {/* Summary */}
            <div className="grid grid-cols-5 gap-4 mb-6">
              <div className="bg-green-50 p-4 rounded-lg">
                <p className="text-xs text-green-700">Current (0-30)</p>
                <p className="text-lg font-bold text-green-900">{formatCurrency(reportData.total_0_30)}</p>
              </div>
              <div className="bg-yellow-50 p-4 rounded-lg">
                <p className="text-xs text-yellow-700">31-60 Days</p>
                <p className="text-lg font-bold text-yellow-900">{formatCurrency(reportData.total_31_60)}</p>
              </div>
              <div className="bg-orange-50 p-4 rounded-lg">
                <p className="text-xs text-orange-700">61-90 Days</p>
                <p className="text-lg font-bold text-orange-900">{formatCurrency(reportData.total_61_90)}</p>
              </div>
              <div className="bg-red-50 p-4 rounded-lg">
                <p className="text-xs text-red-700">90+ Days</p>
                <p className="text-lg font-bold text-red-900">{formatCurrency(reportData.total_90_plus)}</p>
              </div>
              <div className="bg-blue-50 p-4 rounded-lg">
                <p className="text-xs text-blue-700">Total</p>
                <p className="text-lg font-bold text-blue-900">{formatCurrency(reportData.total_amount)}</p>
              </div>
            </div>

            {/* Detail Table */}
            <table className="w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="py-2 px-4 text-left">Party</th>
                  <th className="py-2 px-4 text-left">Invoice</th>
                  <th className="py-2 px-4 text-left">Due Date</th>
                  <th className="py-2 px-4 text-right">0-30</th>
                  <th className="py-2 px-4 text-right">31-60</th>
                  <th className="py-2 px-4 text-right">61-90</th>
                  <th className="py-2 px-4 text-right">90+</th>
                  <th className="py-2 px-4 text-right">Total</th>
                </tr>
              </thead>
              <tbody>
                {reportData.items?.map((item: any, idx: number) => (
                  <tr key={idx} className="border-b">
                    <td className="py-2 px-4">{item.party_name}</td>
                    <td className="py-2 px-4">{item.invoice_number}</td>
                    <td className="py-2 px-4">{formatDate(item.due_date)}</td>
                    <td className="py-2 px-4 text-right">{item.amount_0_30 > 0 ? formatCurrency(item.amount_0_30) : '-'}</td>
                    <td className="py-2 px-4 text-right">{item.amount_31_60 > 0 ? formatCurrency(item.amount_31_60) : '-'}</td>
                    <td className="py-2 px-4 text-right">{item.amount_61_90 > 0 ? formatCurrency(item.amount_61_90) : '-'}</td>
                    <td className="py-2 px-4 text-right">{item.amount_90_plus > 0 ? formatCurrency(item.amount_90_plus) : '-'}</td>
                    <td className="py-2 px-4 text-right font-semibold">{formatCurrency(item.outstanding_amount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        );

      default:
        return (
          <div className="text-center py-12">
            <FileText className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600">Report data will be displayed here</p>
          </div>
        );
    }
  };

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Financial Reports</h1>
        <p className="text-gray-600 mt-2">Generate and view financial reports</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {reports.map((report) => {
          const Icon = report.icon;
          return (
            <button
              key={report.name}
              onClick={() => setSelectedReport(report)}
              className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-all text-left group"
            >
              <div className={`w-12 h-12 rounded-lg flex items-center justify-center mb-4 ${getColorClasses(report.color)}`}>
                <Icon className="w-6 h-6" />
              </div>
              <h3 className="font-semibold text-gray-900 mb-2 group-hover:text-emerald-600 transition-colors">
                {report.name}
              </h3>
              <p className="text-sm text-gray-600">{report.description}</p>
              <div className="mt-4 flex items-center gap-2 text-sm text-emerald-600 opacity-0 group-hover:opacity-100 transition-opacity">
                <Download className="w-4 h-4" />
                <span>Generate Report</span>
              </div>
            </button>
          );
        })}
      </div>

      {/* Modal */}
      {selectedReport && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] overflow-hidden flex flex-col">
            {/* Modal Header */}
            <div className="flex items-center justify-between p-6 border-b">
              <div>
                <h2 className="text-2xl font-bold text-gray-900">{selectedReport.name}</h2>
                <p className="text-gray-600 text-sm">{selectedReport.description}</p>
              </div>
              <button
                onClick={closeModal}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <X className="w-6 h-6" />
              </button>
            </div>

            {/* Modal Body */}
            <div className="flex-1 overflow-y-auto p-6">
              {!reportData ? (
                <div className="space-y-6">
                  {/* Date Selection */}
                  {selectedReport.requiresDateRange && (
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          From Date
                        </label>
                        <input
                          type="date"
                          value={fromDate}
                          onChange={(e) => setFromDate(e.target.value)}
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          To Date
                        </label>
                        <input
                          type="date"
                          value={toDate}
                          onChange={(e) => setToDate(e.target.value)}
                          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                        />
                      </div>
                    </div>
                  )}

                  {selectedReport.requiresSingleDate && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        As of Date
                      </label>
                      <input
                        type="date"
                        value={asOfDate}
                        onChange={(e) => setAsOfDate(e.target.value)}
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500"
                      />
                    </div>
                  )}

                  {/* Generate Button */}
                  <button
                    onClick={handleGenerateReport}
                    disabled={loading}
                    className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {loading ? (
                      <>
                        <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                        <span>Generating...</span>
                      </>
                    ) : (
                      <>
                        <Download className="w-5 h-5" />
                        <span>Generate Report</span>
                      </>
                    )}
                  </button>
                </div>
              ) : (
                <div>
                  {renderReportContent()}
                  
                  {/* Action Buttons */}
                  <div className="mt-6 flex gap-3">
                    <button
                      onClick={() => setReportData(null)}
                      className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                    >
                      Generate New Report
                    </button>
                    <button
                      onClick={() => toast.info('PDF export coming soon')}
                      className="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                    >
                      <Download className="w-4 h-4" />
                      <span>Export PDF</span>
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
