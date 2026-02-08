'use client';

import { useState, useEffect } from 'react';
import { Settings, DollarSign, Check, RefreshCw } from 'lucide-react';
import PageHeader from '@/components/PageHeader';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import ToggleSwitch from '@/components/ui/ToggleSwitch';
import SecretInput from '@/components/ui/SecretInput';
import StatusAlert from '@/components/ui/StatusAlert';
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
    trading: { autoStopLoss: true, autoProfitTaking: false },
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

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await apiClient.get('/settings');
      console.log('[Settings] Fetched from server:', response);
      
      if (response) {
        setSettings((prev) => ({
          ...prev,
          trading: {
            autoStopLoss: response.trading?.autoStopLoss ?? prev.trading.autoStopLoss,
            autoProfitTaking: response.trading?.autoProfitTaking ?? prev.trading.autoProfitTaking,
          },
          api: {
            alpacaKey: '',
            alpacaKeyMasked: response.api?.alpacaKeyMasked || '****...****',
            alpacaSecret: '',
            alpacaSecretMasked: response.api?.alpacaSecretMasked || '****...****',
            finnhubKey: '',
            finnhubKeyMasked: response.api?.finnhubKeyMasked || '****...****',
          },
        }));
      }
    } catch (err: any) {
      console.error('[Settings] Fetch error:', err);
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

      Object.keys(payload.api).forEach(
        (key) =>
          payload.api[key as keyof typeof payload.api] === undefined &&
          delete payload.api[key as keyof typeof payload.api]
      );

      console.log('[Settings] Saving settings:', { ...payload, api: '***hidden***' });
      const response = await apiClient.post('/settings', payload);
      console.log('[Settings] Save response:', response);
      
      setSuccess('Settings saved successfully!');
      
      // Refresh settings from server to get masked values
      await fetchSettings();
      
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      console.error('[Settings] Save error:', err);
      setError(err.message || 'Failed to save settings');
    } finally {
      setIsLoading(false);
    }
  };

  const handleTradingChange = (key: keyof TradingSettings, value: boolean) => {
    setSettings((prev) => ({ ...prev, trading: { ...prev.trading, [key]: value } }));
  };

  const handleApiConfigChange = (key: keyof ApiConfig, value: string) => {
    setSettings((prev) => ({ ...prev, api: { ...prev.api, [key]: value } }));
  };

  const formatLastUpdated = (date?: string) =>
    date ? `Last updated: ${new Date(date).toLocaleDateString()}` : 'Never updated';

  return (
    <div className="space-y-6">
      <PageHeader title="Settings" description="Manage your trading preferences and integrations" />

      {error && <StatusAlert message={error} variant="error" />}
      {success && <StatusAlert message={success} variant="success" />}

      <div className="space-y-6">
        {/* Trading Settings */}
        <Card>
          <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
            <DollarSign className="w-5 h-5" />
            Trading Settings
          </h2>
          <div className="space-y-4">
            <ToggleSwitch
              label="Auto Stop Loss"
              description="Automatically set stop losses on new trades"
              checked={settings.trading.autoStopLoss}
              onChange={(v) => handleTradingChange('autoStopLoss', v)}
            />
            <ToggleSwitch
              label="Auto Profit Taking"
              description="Automatically take profits at target levels"
              checked={settings.trading.autoProfitTaking}
              onChange={(v) => handleTradingChange('autoProfitTaking', v)}
            />
          </div>
        </Card>

        {/* API Configuration */}
        <Card>
          <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
            <Settings className="w-5 h-5" />
            API Configuration
          </h2>
          <div className="space-y-4">
            <SecretInput
              label="Alpaca API Key"
              value={settings.api.alpacaKey || ''}
              maskedValue={settings.api.alpacaKeyMasked}
              onChange={(v) => handleApiConfigChange('alpacaKey', v)}
              placeholder="Enter your Alpaca API key"
              subtext={formatLastUpdated(settings.api.alpacaKeyLastUpdated)}
            />
            <SecretInput
              label="Alpaca API Secret"
              value={settings.api.alpacaSecret || ''}
              maskedValue={settings.api.alpacaSecretMasked}
              onChange={(v) => handleApiConfigChange('alpacaSecret', v)}
              placeholder="Enter your Alpaca API secret"
              subtext={formatLastUpdated(settings.api.alpacaSecretLastUpdated)}
            />
            <SecretInput
              label="Finnhub API Key"
              value={settings.api.finnhubKey || ''}
              maskedValue={settings.api.finnhubKeyMasked}
              onChange={(v) => handleApiConfigChange('finnhubKey', v)}
              placeholder="Enter your Finnhub API key"
              subtext={
                <>
                  Used for news scraping and market data. Get your key at{' '}
                  <a href="https://finnhub.io" target="_blank" rel="noopener noreferrer" className="text-blue-400 hover:text-blue-300">
                    finnhub.io
                  </a>
                </>
              }
            />
          </div>
        </Card>

        {/* Actions */}
        <div className="flex justify-end gap-4">
          <Button variant="secondary" icon={<RefreshCw className="w-4 h-4" />} onClick={fetchSettings}>
            Reset
          </Button>
          <Button icon={<Check className="w-4 h-4" />} onClick={saveSettings} loading={isLoading}>
            Save Settings
          </Button>
        </div>
      </div>
    </div>
  );
}
