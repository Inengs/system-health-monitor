import { useEffect, useState } from "react";
import { GetConfig, SaveConfig } from "../wailsjs/go/main/App";

type Config = {
  geminiApiKey: string;
};

export default function Settings({ onClose }: { onClose: () => void }) {
  const [config, setConfig] = useState<Config>({ geminiApiKey: "" });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [status, setStatus] = useState<string | null>(null);

  useEffect(() => {
    GetConfig().then((cfg: any) => {
      setConfig(cfg);
      setLoading(false);
    });
  }, []);

  const handleSave = async () => {
    setSaving(true);
    setStatus(null);
    const errMsg = await SaveConfig(config);
    setSaving(false);
    setStatus(errMsg || "Saved.");
  };

  if (loading) return <div className="settings">Loading settings…</div>;

  return (
    <div className="settings">
      <div className="settings-header">
        <h2>Settings</h2>
        <button onClick={onClose}>Close</button>
      </div>

      <label htmlFor="gemini-key">
        Gemini API key <span className="optional">(optional)</span>
      </label>
      <p className="settings-hint">
        Enables AI-generated plain-language suggestions when something is
        using a lot of CPU or memory. Get a free key at{" "}
        <a href="https://aistudio.google.com" target="_blank" rel="noreferrer">
          aistudio.google.com
        </a>
        . Leave blank to use rule-based alerts only.
      </p>
      <input
        id="gemini-key"
        type="password"
        value={config.geminiApiKey}
        onChange={(e) => setConfig({ ...config, geminiApiKey: e.target.value })}
        placeholder="AIza..."
      />

      <div className="settings-actions">
        <button onClick={handleSave} disabled={saving}>
          {saving ? "Saving…" : "Save"}
        </button>
        {status && <span className="settings-status">{status}</span>}
      </div>
    </div>
  );
}