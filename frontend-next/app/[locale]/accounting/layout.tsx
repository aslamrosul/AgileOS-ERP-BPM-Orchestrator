'use client';

import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { 
  LayoutDashboard,
  BookOpen,
  FileText,
  Users,
  ShoppingCart,
  TrendingUp,
  DollarSign,
  PieChart,
  Settings,
  ChevronRight
} from 'lucide-react';

const navigation = [
  {
    name: 'Dashboard',
    href: '/en/accounting',
    icon: LayoutDashboard,
    description: 'Overview & Analytics'
  },
  {
    name: 'Chart of Accounts',
    href: '/en/accounting/chart-of-accounts',
    icon: BookOpen,
    description: 'Account Structure'
  },
  {
    name: 'Journal Entries',
    href: '/en/accounting/journal-entries',
    icon: FileText,
    description: 'General Ledger'
  },
  {
    name: 'Vendors',
    href: '/en/accounting/vendors',
    icon: Users,
    description: 'Supplier Management'
  },
  {
    name: 'Customers',
    href: '/en/accounting/customers',
    icon: ShoppingCart,
    description: 'Customer Management'
  },
  {
    name: 'Invoices',
    href: '/en/accounting/invoices',
    icon: FileText,
    description: 'AP & AR'
  },
  {
    name: 'Payments',
    href: '/en/accounting/payments',
    icon: DollarSign,
    description: 'Payment Transactions'
  },
  {
    name: 'Reports',
    href: '/en/accounting/reports',
    icon: PieChart,
    description: 'Financial Reports'
  },
  {
    name: 'Budget',
    href: '/en/accounting/budget',
    icon: TrendingUp,
    description: 'Budget Management'
  },
  {
    name: 'Settings',
    href: '/en/accounting/settings',
    icon: Settings,
    description: 'Configuration'
  }
];

export default function AccountingLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();

  const isActive = (href: string) => {
    if (href === '/en/accounting') {
      return pathname === href;
    }
    return pathname?.startsWith(href);
  };

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="w-72 bg-white border-r border-gray-200 flex flex-col">
        {/* Header */}
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-lg flex items-center justify-center">
              <DollarSign className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-lg font-bold text-gray-900">Accounting</h1>
              <p className="text-xs text-gray-500">Financial Management</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto p-4">
          <div className="space-y-1">
            {navigation.map((item) => {
              const Icon = item.icon;
              const active = isActive(item.href);
              
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={`
                    group flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200
                    ${active 
                      ? 'bg-emerald-50 text-emerald-700 shadow-sm' 
                      : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                    }
                  `}
                >
                  <Icon className={`w-5 h-5 flex-shrink-0 ${active ? 'text-emerald-600' : 'text-gray-400 group-hover:text-gray-600'}`} />
                  <div className="flex-1 min-w-0">
                    <p className={`text-sm font-medium truncate ${active ? 'text-emerald-700' : ''}`}>
                      {item.name}
                    </p>
                    <p className="text-xs text-gray-500 truncate">
                      {item.description}
                    </p>
                  </div>
                  {active && (
                    <ChevronRight className="w-4 h-4 text-emerald-600" />
                  )}
                </Link>
              );
            })}
          </div>
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-gray-200">
          <div className="bg-gradient-to-br from-emerald-50 to-teal-50 rounded-lg p-4">
            <div className="flex items-center gap-2 mb-2">
              <div className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></div>
              <p className="text-xs font-medium text-emerald-900">System Status</p>
            </div>
            <p className="text-xs text-emerald-700">All systems operational</p>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 overflow-y-auto">
        {children}
      </main>
    </div>
  );
}
