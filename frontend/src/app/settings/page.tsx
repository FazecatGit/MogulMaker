'use client';

import { useState, useEffect } from 'react';
import {
  Settings,
  DollarSign,
  AlertCircle,
  Copy,
  Check,
  RefreshCw,
  Eye,
  EyeOff,
} from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import apiClient from '@/lib/apiClient';

interface ApiConfig {
  alpacaKey?: string;
  alpacaKeyLastUpdated?: string;
  alpacaKeyMasked?: string;
  alpacaSecret?: string;
  alpacaSecretLastUpdated?: string;
  alpacaSecretMasked?: string;
  finnhubKey?: string;
  finnhubKeyLastUpdated?: string;
  finnhubKeyMasked?: string;
}

interface TradingSettings {
  autoStopLoss: boolean;
  autoProfitTaking: boolean;
}

interface SettingsData {
  trading: TradingSettings;
  api: ApiConfig;
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<SettingsData>({
    trading: {
      autoStopLoss: true,
      autoProfitTaking: false,
    },
    api: {
      alpacaKey: '',
      alpacaKeyMasked: '****...****',
      alpacaSecret: '',
      alpacaSecretMasked: '****...****',
      finnhubKey: '',
      finnhubKeyMasked: '****...****',
    },
  });

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [showApiKey, setShowApiKey] = useState(false);
  const [showApiSecret, setShowApiSecret] = useState(false);
  const [showFinnhubKey, setShowFinnhubKey] = useState(false);
  const [copiedField, setCopiedField] = useState<string | null>(null);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.get('/api/settings');
      if (response.data) {
        setSettings((prev) => ({
          ...prev,
          trading: {
            autoStopLoss: response.data.trading?.autoStopLoss ?? prev.trading.autoStopLoss,
            autoProfitTaking: response.data.trading?.autoProfitTaking ?? prev.trading.autoProfitTaking,
          },
          api: {
            alpacaKey: '',
            alpacaKeyMasked: response.data.api?.alpacaKeyMasked || '****...****',
            alpacaSecret: '',
            alpacaSecretMasked: response.data.api?.alpacaSecretMasked || '****...****',
            finnhubKey: '',
            finnhubKeyMasked: response.data.api?.finnhubKeyMasked || '****...****',
          },
        }));
      }
    } catch (err: any) {
      setError(err.message || 'Failed to fetch settings');
    } finally {
      setIsLoading(false);
    }
  };

  const saveSettings = async () => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const payload = {
        trading: settings.trading,
        api: {
          alpacaKey: settings.api.alpacaKey || undefined,
          alpacaSecret: settings.api.alpacaSecret || undefined,
          finnhubKey: settings.api.finnhubKey || undefined,
        },
      };

      // Remove undefined values
      Object.keys(payload.api).forEach(
        (key) => payload.api[key as keyof typeof payload.api] === undefined && delete payload.api[key as keyof typeof payload.api]
      );

      await apiClient.post('/api/settings', payload);
      setSuccess('Settings saved successfully!');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err.message || 'Failed to save settings');
    } finally {
      setIsLoading(false);
    }
  };

  const handleTradingChange = (key: keyof TradingSettings, value: any) => {
    setSettings((prev) => ({
      ...prev,
      trading: { ...prev.trading, [key]: value },
    }));
  };

  const handleApiConfigChange = (key: keyof ApiConfig, value: string) => {
    setSettings((prev) => ({
      ...prev,
      api: { ...prev.api, [key]: value },
    }));
  };

  const copyToClipboard = (value: string, field: string) => {
    navigator.clipboard.writeText(value);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <PageHeader title="Settings" description="Manage your trading preferences and integrations" />

      {/* Status Messages */}
      {error && (
        <div className="bg-red-500/20 border border-red-500/50 rounded-lg p-4 text-red-400 flex items-center gap-2">
          <AlertCircle className="w-5 h-5" />
          {error}
        </div>
      )}

      {success && (
        <div className="bg-green-500/20 border border-green-500/50 rounded-lg p-4 text-green-400 flex items-center gap-2">
          <Check className="w-5 h-5" />
          {success}
        </div>
      )}

      <div className="space-y-6">
        {/* Trading Settings Section */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
            <DollarSign className="w-5 h-5" />
            Trading Settings
          </h2>
          <div className="space-y-4">
            {/* Auto Stop Loss */}
            <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
              <div>
                <p className="font-semibold text-white">Auto Stop Loss</p>
                <p className="text-xs text-slate-400">Automatically set stop losses on new trades</p>
              </div>
              <button
                onClick={() =>
                  handleTradingChange('autoStopLoss', !settings.trading.autoStopLoss)
                }
                className={`relative w-14 h-8 rounded-full transition ${
                  settings.trading.autoStopLoss ? 'bg-blue-600' : 'bg-slate-600'
                }`}
              >
                <div
                  className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                    settings.trading.autoStopLoss ? 'translate-x-6' : ''
                  }`}
                ></div>
              </button>
            </div>

            {/* Auto Profit Taking */}
            <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
              <div>
                <p className="font-semibold text-white">Auto Profit Taking</p>
                <p className="text-xs text-slate-400">Automatically take profits at target levels</p>
              </div>
              <button
                onClick={() =>
                  handleTradingChange('autoProfitTaking', !settings.trading.autoProfitTaking)
                }
                className={`relative w-14 h-8 rounded-full transition ${
                  settings.trading.autoProfitTaking ? 'bg-blue-600' : 'bg-slate-600'
                }`}
              >
                <div
                  className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                    settings.trading.autoProfitTaking ? 'translate-x-6' : ''
                  }`}
                ></div>
              </button>
            </div>
          </div>
        </div>

        {/* API Configuration Section */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
            <Settings className="w-5 h-5" />
            API Configuration
          </h2>

          <div className="space-y-4">
            {/* Alpaca API Key */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Alpaca API Key
              </label>
              <div className="flex gap-2">
                <div className="flex-1 relative">
                  <input
                    type={showApiKey ? 'text' : 'password'}
                    value={settings.api.alpacaKey || settings.api.alpacaKeyMasked}
                    onChange={(e) => handleApiConfigChange('alpacaKey', e.target.value)}
                    placeholder="Enter your Alpaca API key"
                    className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
                  />
                  <button
                    onClick={() => setShowApiKey(!showApiKey)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-slate-400 hover:text-slate-300"
                  >
                    {showApiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <button
                  onClick={() =>
                    copyToClipboard(
                      settings.api.alpacaKey || '',
                      'alpaca-key'
                    )
                  }
                  className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded transition flex items-center gap-2"
                >
                  {copiedField === 'alpaca-key' ? (
                    <Check className="w-4 h-4" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </button>
              </div>
              <p className="text-xs text-slate-400 mt-1">
                {settings.api.alpacaKeyLastUpdated
                  ? `Last updated: ${new Date(settings.api.alpacaKeyLastUpdated).toLocaleDateString()}`
                  : 'Never updated'}
              </p>
            </div>

            {/* Alpaca API Secret */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Alpaca API Secret
              </label>
              <div className="flex gap-2">
                <div className="flex-1 relative">
                  <input
                    type={showApiSecret ? 'text' : 'password'}
                    value={settings.api.alpacaSecret || settings.api.alpacaSecretMasked}
                    onChange={(e) => handleApiConfigChange('alpacaSecret', e.target.value)}
                    placeholder="Enter your Alpaca API secret"
                    className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
                  />
                  <button
                    onClick={() => setShowApiSecret(!showApiSecret)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-slate-400 hover:text-slate-300"
                  >
                    {showApiSecret ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <button
                  onClick={() =>
                    copyToClipboard(
                      settings.api.alpacaSecret || '',
                      'alpaca-secret'
                    )
                  }
                  className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded transition flex items-center gap-2"
                >
                  {copiedField === 'alpaca-secret' ? (
                    <Check className="w-4 h-4" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </button>
              </div>
              <p className="text-xs text-slate-400 mt-1">
                {settings.api.alpacaSecretLastUpdated
                  ? `Last updated: ${new Date(settings.api.alpacaSecretLastUpdated).toLocaleDateString()}`
                  : 'Never updated'}
              </p>
            </div>

            {/* Finnhub API Key */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Finnhub API Key
              </label>
              <div className="flex gap-2">
                <div className="flex-1 relative">
                  <input
                    type={showFinnhubKey ? 'text' : 'password'}
                    value={settings.api.finnhubKey || settings.api.finnhubKeyMasked}
                    onChange={(e) => handleApiConfigChange('finnhubKey', e.target.value)}
                    placeholder="Enter your Finnhub API key"
                    className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
                  />
                  <button
                    onClick={() => setShowFinnhubKey(!showFinnhubKey)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-slate-400 hover:text-slate-300"
                  >
                    {showFinnhubKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
                <button
                  onClick={() =>
                    copyToClipboard(
                      settings.api.finnhubKey || '',
                      'finnhub-key'
                    )
                  }
                  className="px-4 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded transition flex items-center gap-2"
                >
                  {copiedField === 'finnhub-key' ? (
                    <Check className="w-4 h-4" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </button>
              </div>
              <p className="text-xs text-slate-400 mt-1">
                Used for news scraping and market data. Get your key at{' '}
                <a href="https://finnhub.io" target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:text-blue-300">
                  finnhub.io
                </a>
              </p>
            </div>
          </div>
        </div>

        {/* Save Button */}
        <div className="flex justify-end gap-4">
          <button
            onClick={() => fetchSettings()}
            className="px-6 py-2 bg-slate-700 hover:bg-slate-600 text-white rounded-lg font-semibold transition flex items-center gap-2"
          >
            <RefreshCw className="w-4 h-4" />
            Reset
          </button>
          <button
            onClick={saveSettings}
            disabled={isLoading}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-slate-700 text-white rounded-lg font-semibold transition flex items-center gap-2"
          >
            <Check className="w-4 h-4" />
            Save Settings
          </button>
        </div>
      </div>
    </div>
  );
}
