"use client";

import { useState, useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { Globe } from "lucide-react";
import { locales, localeNames, type Locale } from "@/i18n";

export default function LanguageSwitcher() {
  const router = useRouter();
  const pathname = usePathname();
  const [currentLocale, setCurrentLocale] = useState<Locale>("en");
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    // Get current locale from pathname or localStorage
    const pathLocale = pathname.split("/")[1] as Locale;
    if (locales.includes(pathLocale)) {
      setCurrentLocale(pathLocale);
    } else {
      // Check localStorage
      const savedLocale = localStorage.getItem("locale") as Locale;
      if (savedLocale && locales.includes(savedLocale)) {
        setCurrentLocale(savedLocale);
      }
    }
  }, [pathname]);

  const switchLanguage = (locale: Locale) => {
    // Save to localStorage
    localStorage.setItem("locale", locale);
    
    // Update cookie for server-side
    document.cookie = `NEXT_LOCALE=${locale}; path=/; max-age=31536000`;
    
    // Get current path without locale
    const segments = pathname.split("/").filter(Boolean);
    const pathWithoutLocale = segments.slice(1).join("/");
    
    // Navigate to new locale
    router.push(`/${locale}/${pathWithoutLocale}`);
    setIsOpen(false);
  };

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-gray-100 transition-colors"
        title="Change Language"
      >
        <Globe className="w-5 h-5 text-gray-600" />
        <span className="text-sm font-medium text-gray-700">
          {localeNames[currentLocale]}
        </span>
      </button>

      {isOpen && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-40"
            onClick={() => setIsOpen(false)}
          />

          {/* Dropdown */}
          <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
            {locales.map((locale) => (
              <button
                key={locale}
                onClick={() => switchLanguage(locale)}
                className={`w-full px-4 py-2 text-left text-sm hover:bg-gray-100 transition-colors ${
                  currentLocale === locale
                    ? "bg-indigo-50 text-indigo-600 font-medium"
                    : "text-gray-700"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span>{localeNames[locale]}</span>
                  {currentLocale === locale && (
                    <span className="text-indigo-600">✓</span>
                  )}
                </div>
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
