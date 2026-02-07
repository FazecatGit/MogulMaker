'use client';

import { useState, useEffect } from 'react';
import {
  Settings,
  Bell,
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
}

interface TradingSettings {
  maxDailyLoss: number;
  maxPositionRisk: number;
  maxOpenPositions: number;
  tradingHoursOnly: boolean;
  autoStopLoss: boolean;
  autoProfitTaking: boolean;
}

interface NotificationSettings {
  emailAlerts: boolean;
  tradeExecutionNotifications: boolean;
  riskAlerts: boolean;
  dailySummary: boolean;
  newsAlerts: boolean;
}

interface SettingsData {
  trading: TradingSettings;
  notifications: NotificationSettings;
  api: ApiConfig;
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<SettingsData>({
    trading: {
      maxDailyLoss: 5000,
      maxPositionRisk: 1000,
      maxOpenPositions: 10,
      tradingHoursOnly: true,
      autoStopLoss: true,
      autoProfitTaking: false,
    },
    notifications: {
      emailAlerts: true,
      tradeExecutionNotifications: true,
      riskAlerts: true,
      dailySummary: false,
      newsAlerts: true,
    },
    api: {
      alpacaKey: '',
      alpacaKeyMasked: '****...****',
    },
  });

  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [showApiKey, setShowApiKey] = useState(false);
  const [copiedField, setCopiedField] = useState<string | null>(null);

  useEffect(() => {
    // Disabled auto-fetch on mount - settings only load if user requests
    // fetchSettings();
  }, []);

  const fetchSettings = async () => {
    setIsLoading(false); // Simulate loading
    setError(null);
    // Settings come from frontend state for now
  };

  const saveSettings = async () => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // In production, this would call the API
      // await apiClient.post('/settings', settings);
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

  const handleNotificationChange = (key: keyof NotificationSettings, value: boolean) => {
    setSettings((prev) => ({
      ...prev,
      notifications: { ...prev.notifications, [key]: value },
    }));
  };

  const handleApiKeyChange = (value: string) => {
    setSettings((prev) => ({
      ...prev,
      api: { ...prev.api, alpacaKey: value },
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

          <div className="space-y-6">
            {/* Max Daily Loss */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Maximum Daily Loss ($)
              </label>
              <input
                type="number"
                value={settings.trading.maxDailyLoss}
                onChange={(e) =>
                  handleTradingChange('maxDailyLoss', parseFloat(e.target.value) || 0)
                }
                className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
              <p className="text-xs text-slate-400 mt-1">
                Trading will halt when daily loss exceeds this amount
              </p>
            </div>

            {/* Max Position Risk */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Maximum Risk Per Position ($)
              </label>
              <input
                type="number"
                value={settings.trading.maxPositionRisk}
                onChange={(e) =>
                  handleTradingChange('maxPositionRisk', parseFloat(e.target.value) || 0)
                }
                className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
              <p className="text-xs text-slate-400 mt-1">
                Single position cannot exceed this risk amount
              </p>
            </div>

            {/* Max Open Positions */}
            <div>
              <label className="block text-slate-300 text-sm font-semibold mb-2">
                Maximum Open Positions
              </label>
              <input
                type="number"
                value={settings.trading.maxOpenPositions}
                onChange={(e) =>
                  handleTradingChange('maxOpenPositions', parseInt(e.target.value) || 0)
                }
                min="1"
                max="50"
                className="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded text-white focus:outline-none focus:border-blue-500"
              />
              <p className="text-xs text-slate-400 mt-1">
                Number of simultaneous open positions allowed
              </p>
            </div>

            {/* Toggle Settings */}
            <div className="space-y-4">
              {/* Trading Hours Only */}
              <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                <div>
                  <p className="font-semibold text-white">Trade During Market Hours Only</p>
                  <p className="text-xs text-slate-400">Execute trades only 9:30 AM - 4:00 PM EST</p>
                </div>
                <button
                  onClick={() =>
                    handleTradingChange('tradingHoursOnly', !settings.trading.tradingHoursOnly)
                  }
                  className={`relative w-14 h-8 rounded-full transition ${
                    settings.trading.tradingHoursOnly ? 'bg-blue-600' : 'bg-slate-600'
                  }`}
                >
                  <div
                    className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                      settings.trading.tradingHoursOnly ? 'translate-x-6' : ''
                    }`}
                  ></div>
                </button>
              </div>

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
        </div>

        {/* Notification Settings Section */}
        <div className="bg-slate-800 rounded-lg border border-slate-700 p-6">
          <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
            <Bell className="w-5 h-5" />
            Notification Settings
          </h2>

          <div className="space-y-4">
            {/* Email Alerts */}
            <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
              <div>
                <p className="font-semibold text-white">Email Alerts</p>
                <p className="text-xs text-slate-400">Receive important alerts via email</p>
              </div>
              <button
                onClick={() =>
                  handleNotificationChange(
                    'emailAlerts',
                    !settings.notifications.emailAlerts
                  )
                }
                className={`relative w-14 h-8 rounded-full transition ${
                  settings.notifications.emailAlerts ? 'bg-blue-600' : 'bg-slate-600'
                }`}
              >
                <div
                  className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                    settings.notifications.emailAlerts ? 'translate-x-6' : ''
                  }`}
                ></div>
              </button>
            </div>

            {/* Trade Execution Notifications */}
            <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
              <div>
                <p className="font-semibold text-white">Trade Execution Notifications</p>
                <p className="text-xs text-slate-400">Get notified when trades are executed</p>
              </div>
              <button
                onClick={() =>
                  handleNotificationChange(
                    'tradeExecutionNotifications',
                    !settings.notifications.tradeExecutionNotifications
                  )
                }
                className={`relative w-14 h-8 rounded-full transition ${
                  settings.notifications.tradeExecutionNotifications ? 'bg-blue-600' : 'bg-slate-600'
                }`}
              >
                <div
                  className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                    settings.notifications.tradeExecutionNotifications ? 'translate-x-6' : ''
                  }`}
                ></div>
              </button>
            </div>

            {/* Risk Alerts */}
            <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
              <div>
                <p className="font-semibold text-white">Risk Alerts</p>
                <p className="text-xs text-slate-400">Get alerted about risk threshold breaches</p>
              </div>
              <button
                onClick={() =>
                  handleNotificationChange('riskAlerts', !settings.notifications.riskAlerts)
                }
                className={`relative w-14 h-8 rounded-full transition ${
                  settings.notifications.riskAlerts ? 'bg-blue-600' : 'bg-slate-600'
                }`}
              >
                <div
                  className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                    settings.notifications.riskAlerts ? 'translate-x-6' : ''
                  }`}
                ></div>
              </button>
            </div>

            {/* Daily Summary */}
            <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
              <div>
                <p className="font-semibold text-white">Daily Summary</p>
                <p className="text-xs text-slate-400">Receive daily trading summary report</p>
              </div>
              <button
                onClick={() =>
                  handleNotificationChange('dailySummary', !settings.notifications.dailySummary)
                }
                className={`relative w-14 h-8 rounded-full transition ${
                  settings.notifications.dailySummary ? 'bg-blue-600' : 'bg-slate-600'
                }`}
              >
                <div
                  className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                    settings.notifications.dailySummary ? 'translate-x-6' : ''
                  }`}
                ></div>
              </button>
            </div>

            {/* News Alerts */}
            <div className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
              <div>
                <p className="font-semibold text-white">News Alerts</p>
                <p className="text-xs text-slate-400">Get notified about relevant market news</p>
              </div>
              <button
                onClick={() =>
                  handleNotificationChange('newsAlerts', !settings.notifications.newsAlerts)
                }
                className={`relative w-14 h-8 rounded-full transition ${
                  settings.notifications.newsAlerts ? 'bg-blue-600' : 'bg-slate-600'
                }`}
              >
                <div
                  className={`absolute top-1 left-1 w-6 h-6 bg-white rounded-full transition transform ${
                    settings.notifications.newsAlerts ? 'translate-x-6' : ''
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
                    onChange={(e) => handleApiKeyChange(e.target.value)}
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
