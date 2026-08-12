import { useEffect, useState } from "react";
import { GetSnapshot, RequestClose } from "../wailsjs/go/main/App";
import "./App.css";

type ProcInfo = {
  pid: number;
  name: string;
  friendly: string;
  cpuPercent: number;
  memPercent: number;
  safeToClose: boolean;
};

type Alert = {
  pid: number;
  name: string;
  reason: string;
  time: string;
};

type Snapshot = {
  topByCpu: ProcInfo[];
  topByMem: ProcInfo[];
  alerts: Alert[];
};

export default function App() {
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const [confirming, setConfirming] = useState<ProcInfo | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  useEffect(() => {
    const poll = async () => setSnapshot(await GetSnapshot());
    poll();
    const id = setInterval(poll, 3000);
    return () => clearInterval(id);
  }, []);

  if (!snapshot) return <div id="App">Loading…</div>;

  const topOffender = snapshot.topByCpu?.[0];

  async function handleConfirmClose() {
    if (!confirming) return;
    const result = await RequestClose(confirming.pid);
    setStatus(result);
    setConfirming(null);
  }

  return (
    <div id="App">
      {topOffender && topOffender.cpuPercent > 80 && (
        <div className="banner">
          {topOffender.friendly} is using {Math.round(topOffender.cpuPercent)}%
          of your CPU — this is probably why things feel slow.
        </div>
      )}

      {status && <div className="status">{status}</div>}

      <ul className="proc-list">
        {snapshot.topByCpu.map((p) => (
          <li key={p.pid}>
            <span className="proc-name">{p.friendly}</span>
            <span className="proc-stat">{Math.round(p.cpuPercent)}% CPU</span>
            {p.safeToClose && (
              <button onClick={() => setConfirming(p)}>Close</button>
            )}
          </li>
        ))}
      </ul>

      {confirming && (
        <div className="modal">
          <p>Close {confirming.friendly}? Unsaved work in it will be lost.</p>
          <button onClick={handleConfirmClose}>Yes, close it</button>
          <button onClick={() => setConfirming(null)}>Cancel</button>
        </div>
      )}
    </div>
  );
}
