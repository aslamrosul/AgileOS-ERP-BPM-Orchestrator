// Backup of original home page
"use client";

import Link from "next/link";
import { useTranslations } from 'next-intl';
import { Workflow, Zap, Shield, BarChart3 } from "lucide-react";
import { ConnectionStatus } from "@/components/ConnectionStatus";
import { NotificationsPanel } from "@/components/NotificationsPanel";
import LanguageSwitcher from "@/components/LanguageSwitcher";

export default function Home() {
  const t = useTranslations('common');
  const tHome = useTranslations('home');

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Header with WebSocket Status and Language Switcher */}
      <div className="absolute top-4 right-4 flex items-center space-x-4">
        <LanguageSwitcher />
        <ConnectionStatus />
        <NotificationsPanel />
      </div>
      
      <div className="container mx-auto px-4 py-16">
        <div className="text-center mb-16">
          <h1 className="text-6xl font-bold text-gray-900 mb-4">
            AgileOS
          </h1>
          <p className="text-xl text-gray-600 mb-8">
            {tHome('subtitle')}
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/login"
              className="inline-flex items-center gap-2 bg-blue-600 text-white px-8 py-4 rounded-lg text-lg font-semibold hover:bg-blue-700 transition-colors"
            >
              {t('login')}
            </Link>
            <Link
              href="/workflow"
              className="inline-flex items-center gap-2 bg-indigo-600 text-white px-8 py-4 rounded-lg text-lg font-semibold hover:bg-indigo-700 transition-colors"
            >
              <Workflow className="w-6 h-6" />
              {tHome('openWorkflowBuilder')}
            </Link>
            <Link
              href="/analytics"
              className="inline-flex items-center gap-2 bg-purple-600 text-white px-8 py-4 rounded-lg text-lg font-semibold hover:bg-purple-700 transition-colors"
            >
              <BarChart3 className="w-6 h-6" />
              {tHome('viewAnalytics')}
            </Link>
          </div>
        </div>

        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          <div className="bg-white p-6 rounded-xl shadow-lg">
            <div className="w-12 h-12 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
              <Zap className="w-6 h-6 text-indigo-600" />
            </div>
            <h3 className="text-xl font-semibold mb-2">{tHome('feature1Title')}</h3>
            <p className="text-gray-600">
              {tHome('feature1Description')}
            </p>
          </div>

          <div className="bg-white p-6 rounded-xl shadow-lg">
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
              <Shield className="w-6 h-6 text-green-600" />
            </div>
            <h3 className="text-xl font-semibold mb-2">{tHome('feature2Title')}</h3>
            <p className="text-gray-600">
              {tHome('feature2Description')}
            </p>
          </div>

          <div className="bg-white p-6 rounded-xl shadow-lg">
            <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
              <BarChart3 className="w-6 h-6 text-purple-600" />
            </div>
            <h3 className="text-xl font-semibold mb-2">{tHome('feature3Title')}</h3>
            <p className="text-gray-600">
              {tHome('feature3Description')}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
