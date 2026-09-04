import React, {FormEvent, ReactNode, useEffect, useState} from "react";
import {createRoot} from "react-dom/client";
import {MeshAlotBrand, MeshHeroArtwork} from "./brand";
import "./style.css";

type User = {id: string; email: string};
type NodeRecord = {node_id: string; status: string; mode: string; agent_version: string; last_heartbeat?: string};
type EnrollmentCode = {enrollment_code: string; expires_at: string};
type Dashboard = {
  balance_microunits: number;
  online_nodes: number;
  compute_score: number | null;
  network_score: number | null;
  current_status: string;
  current_rate_microunits: number | null;
  today_earnings_microunits: number;
};
type WalletTransaction = {id: string; job_id?: string; transaction_type: string; amount_microunits: number; created_at: string};
type Wallet = {balance_microunits: number; transactions: WalletTransaction[]};
type Job = {id: string; provider_node_id?: string; status: string; created_at: string};

const api = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");

class ApiError extends Error { constructor(message: string, public status = 0) { super(message); } }

async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${api}${path}`, {
      ...options,
      credentials: "include",
      headers: {"Content-Type": "application/json", ...(options.headers ?? {})},
    });
  } catch {
    throw new ApiError("MeshAlot is temporarily unreachable. Check your connection and try again.");
  }
  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    try { const body = await response.json(); if (body?.error) message = body.error; } catch { /* keep status message */ }
    throw new ApiError(message, response.status);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

function money(microunits: number | null | undefined) {
  if (microunits == null) return "Pending";
  return `$${(microunits / 1_000_000).toFixed(2)}`;
}

function useRoute() {
  const [route, setRoute] = useState(window.location.pathname || "/dashboard");
  useEffect(() => {
    const onPopState = () => setRoute(window.location.pathname || "/dashboard");
    window.addEventListener("popstate", onPopState);
    return () => window.removeEventListener("popstate", onPopState);
  }, []);
  const navigate = (next: string) => { window.history.pushState({}, "", next); setRoute(next); };
  return {route, navigate};
}

function ErrorBox({message}: {message: string}) { return <div className="error-box" role="alert">{message}</div>; }
function Empty({children}: {children: ReactNode}) { return <div className="empty">{children}</div>; }

function SignIn({onSignedIn, initialError}: {onSignedIn: (user: User) => void; initialError?: string}) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState(initialError ?? "");
  const [busy, setBusy] = useState(false);
  async function submit(event: FormEvent) {
    event.preventDefault(); setError(""); setBusy(true);
    try {
      const session = await apiFetch<{user: User}>("/v1/auth/login", {method: "POST", body: JSON.stringify({email, password})});
      setPassword(""); onSignedIn(session.user);
    } catch (err) { setError(err instanceof Error ? err.message : "Sign in failed"); }
    finally { setBusy(false); }
  }
  return <main className="signin-shell">
    <section className="signin-visual">
      <div className="signin-visual-content">
        <MeshHeroArtwork />
        <div className="signin-visual-copy">
          <p className="signin-kicker">DISTRIBUTED COMPUTE NETWORK</p>
          <h2>Turn independent local machines into one coordinated mesh network.</h2>
          <p>Securely attach computers, see network state, and build toward a marketplace for shared compute.</p>
        </div>
      </div>
    </section>
    <section className="signin-card">
      <p className="eyebrow">MESHALOT POC</p>
      <h1>Sign in to your compute network.</h1>
      <p className="muted">Manage nodes, simulated earnings, and marketplace activity from one account.</p>
      {error && <ErrorBox message={error} />}
      <form onSubmit={submit}>
        <label>Email<input type="email" autoComplete="username" value={email} onChange={e => setEmail(e.target.value)} required /></label>
        <label>Password<input type="password" autoComplete="current-password" value={password} onChange={e => setPassword(e.target.value)} required /></label>
        <button className="primary" type="submit" disabled={busy}>{busy ? "Signing in…" : "Sign in"}</button>
      </form>
    </section>
  </main>;
}

const nav = [
  ["/dashboard", "Dashboard"], ["/nodes", "My Nodes"], ["/marketplace", "Compute Marketplace"],
  ["/run-job", "Run Job"], ["/jobs", "Job History"], ["/wallet", "Wallet"], ["/economics", "Node Economics"],
] as const;

function Shell({user, route, navigate, onLogout, children}: {user: User; route: string; navigate: (path: string) => void; onLogout: () => void; children: ReactNode}) {
  return <div className="app-shell">
    <aside>
      <div><div className="brand"><MeshAlotBrand compact /></div>
        <nav>{nav.map(([path, label]) => <button key={path} className={route === path || (path === "/nodes" && route.startsWith("/nodes/")) ? "active" : ""} onClick={() => navigate(path)}>{label}</button>)}</nav>
      </div>
      <div className="account-block"><span>{user.email}</span><button onClick={onLogout}>Sign out</button></div>
    </aside>
    <main className="content">{children}</main>
  </div>;
}

function DashboardPage() {
  const [data, setData] = useState<Dashboard | null>(null); const [error, setError] = useState("");
  useEffect(() => { apiFetch<Dashboard>("/v1/dashboard").then(setData).catch(e => setError(e.message)); }, []);
  const cards = [
    ["Account balance", data ? money(data.balance_microunits) : "—"], ["Online nodes", data ? String(data.online_nodes) : "—"],
    ["Compute score", data?.compute_score == null ? "Pending" : data.compute_score.toFixed(1)], ["Network score", data?.network_score == null ? "Pending" : data.network_score.toFixed(1)],
    ["Current status", data?.current_status ?? "—"], ["Current rate", data ? money(data.current_rate_microunits) : "—"],
    ["Today's simulated earnings", data ? money(data.today_earnings_microunits) : "—"],
  ];
  return <><PageHeader eyebrow="CONTROL CENTER" title="Dashboard" text="Your MeshAlot POC account at a glance." />{error && <ErrorBox message={error} />}<div className="card-grid">{cards.map(([label, value]) => <section className="metric-card" key={label}><span>{label}</span><strong>{value}</strong></section>)}</div></>;
}

function NodesPage({navigate}: {navigate: (path: string) => void}) {
  const [nodes, setNodes] = useState<NodeRecord[]>([]);
  const [error, setError] = useState("");
  const [codeError, setCodeError] = useState("");
  const [issuing, setIssuing] = useState(false);
  const [issued, setIssued] = useState<EnrollmentCode | null>(null);
  const [copied, setCopied] = useState(false);
  useEffect(() => { apiFetch<{nodes: NodeRecord[]}>("/v1/account/nodes").then(d => setNodes(d.nodes)).catch(e => setError(e.message)); }, []);
  async function generateEnrollmentCode() {
    setCodeError(""); setCopied(false); setIssuing(true);
    try {
      const result = await apiFetch<EnrollmentCode>("/v1/account/enrollment-codes", {method: "POST", body: "{}"});
      setIssued(result);
    } catch (err) {
      setCodeError(err instanceof Error ? err.message : "Unable to create enrollment code");
    } finally {
      setIssuing(false);
    }
  }
  async function copyEnrollmentCode() {
    if (!issued) return;
    try {
      await navigator.clipboard.writeText(issued.enrollment_code);
      setCopied(true);
    } catch {
      setCopied(false);
      setCodeError("Copy was blocked by the browser. Select the enrollment code and copy it manually.");
    }
  }
  return <>
    <PageHeader eyebrow="PROVIDER" title="My Nodes" text="Computers securely attached to this account." />
    <section className="enrollment-card">
      <div className="enrollment-heading">
        <div><p className="eyebrow">SECURE ENROLLMENT</p><h2>Add a computer</h2><p>Create a short-lived, one-time code for the MeshAlot agent. The code is displayed here only and is not saved in this browser.</p></div>
        <button className="primary" type="button" disabled={issuing} onClick={generateEnrollmentCode}>{issuing ? "Generating…" : issued ? "Generate another code" : "Generate enrollment code"}</button>
      </div>
      {codeError && <ErrorBox message={codeError} />}
      {issued && <div className="enrollment-code-panel" aria-live="polite">
        <div className="enrollment-code-label"><span>One-time enrollment code</span><strong>Expires {new Date(issued.expires_at).toLocaleString()}</strong></div>
        <code className="enrollment-code">{issued.enrollment_code}</code>
        <div className="enrollment-actions"><button type="button" onClick={copyEnrollmentCode}>{copied ? "Copied" : "Copy code"}</button><span>Use it once with <code>meshalot-agent enroll</code>. Refreshing or leaving this page removes this displayed copy.</span></div>
      </div>}
    </section>
    {error && <ErrorBox message={error} />}
    {!error && !nodes.length ? <Empty>No nodes are attached to this account yet.</Empty> : <div className="list-card">{nodes.map(node => <button className="list-row" key={node.node_id} onClick={() => navigate(`/nodes/${encodeURIComponent(node.node_id)}`)}><span><strong>{node.node_id}</strong><small>Agent {node.agent_version || "unknown"}</small></span><span className={`status ${node.status}`}>{node.status} · {node.mode}</span></button>)}</div>}
  </>;
}

function NodeDetailsPage({nodeID}: {nodeID: string}) {
  const [node, setNode] = useState<NodeRecord | null>(null); const [error, setError] = useState("");
  useEffect(() => { apiFetch<NodeRecord>(`/v1/account/nodes/${encodeURIComponent(nodeID)}`).then(setNode).catch(e => setError(e.message)); }, [nodeID]);
  return <><PageHeader eyebrow="NODE DETAILS" title={nodeID} text="Identity and current POC connection state." />{error && <ErrorBox message={error} />}{node && <section className="detail-card"><Detail label="Status" value={node.status} /><Detail label="Mode" value={node.mode} /><Detail label="Agent version" value={node.agent_version || "Unknown"} /><Detail label="Last heartbeat" value={node.last_heartbeat ? new Date(node.last_heartbeat).toLocaleString() : "Not yet reported"} /></section>}</>;
}

function WalletPage() {
  const [wallet, setWallet] = useState<Wallet | null>(null); const [error, setError] = useState("");
  useEffect(() => { apiFetch<Wallet>("/v1/account/wallet").then(setWallet).catch(e => setError(e.message)); }, []);
  return <><PageHeader eyebrow="SIMULATED MONEY" title="Wallet" text="Append-only POC earnings and usage ledger." />{error && <ErrorBox message={error} />}{wallet && <section className="hero-balance"><span>Account balance</span><strong>{money(wallet.balance_microunits)}</strong></section>}{wallet && (wallet.transactions.length ? <div className="list-card">{wallet.transactions.map(item => <div className="list-row static" key={item.id}><span><strong>{item.transaction_type}</strong><small>{new Date(item.created_at).toLocaleString()}</small></span><strong>{money(item.amount_microunits)}</strong></div>)}</div> : <Empty>No wallet activity yet.</Empty>)}</>;
}

function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([]); const [error, setError] = useState("");
  useEffect(() => { apiFetch<{jobs: Job[]}>("/v1/account/jobs").then(d => setJobs(d.jobs)).catch(e => setError(e.message)); }, []);
  return <><PageHeader eyebrow="AUDIT TRAIL" title="Job History" text="Jobs consumed by or provided through this account." />{error && <ErrorBox message={error} />}{!error && !jobs.length ? <Empty>No jobs have been recorded for this account.</Empty> : <div className="list-card">{jobs.map(job => <div className="list-row static" key={job.id}><span><strong>{job.id}</strong><small>{new Date(job.created_at).toLocaleString()}</small></span><span className="status">{job.status}</span></div>)}</div>}</>;
}

function PlaceholderPage({eyebrow, title, text, next}: {eyebrow: string; title: string; text: string; next: string}) {
  return <><PageHeader eyebrow={eyebrow} title={title} text={text} /><section className="placeholder"><strong>Dashboard shell ready</strong><p>{next}</p></section></>;
}
function PageHeader({eyebrow, title, text}: {eyebrow: string; title: string; text: string}) { return <header className="page-header"><p className="eyebrow">{eyebrow}</p><h1>{title}</h1><p>{text}</p></header>; }
function Detail({label, value}: {label: string; value: string}) { return <div className="detail"><span>{label}</span><strong>{value}</strong></div>; }

function App() {
  const {route, navigate} = useRoute(); const [user, setUser] = useState<User | null | undefined>(undefined); const [sessionError, setSessionError] = useState("");
  useEffect(() => { apiFetch<{user: User}>("/v1/auth/me").then(s => setUser(s.user)).catch((e: ApiError) => { if (e.status !== 401) setSessionError(e.message); setUser(null); }); }, []);
  if (user === undefined) return <main className="loading">Loading MeshAlot…</main>;
  if (!user) return <SignIn initialError={sessionError} onSignedIn={signedIn => {setSessionError(""); setUser(signedIn); navigate("/dashboard");}} />;
  async function logout() { try { await apiFetch<void>("/v1/auth/logout", {method: "POST"}); } finally { setUser(null); navigate("/dashboard"); } }
  let page: ReactNode;
  if (route === "/dashboard" || route === "/") page = <DashboardPage />;
  else if (route === "/nodes") page = <NodesPage navigate={navigate} />;
  else if (route.startsWith("/nodes/")) page = <NodeDetailsPage nodeID={decodeURIComponent(route.slice("/nodes/".length))} />;
  else if (route === "/marketplace") page = <PlaceholderPage eyebrow="CONSUMER" title="Compute Marketplace" text="Browse eligible compute capacity." next="Marketplace discovery is intentionally a skeleton until node scoring and eligibility milestones are complete." />;
  else if (route === "/run-job") page = <PlaceholderPage eyebrow="CONSUMER" title="Run Job" text="Submit an approved AI workload." next="Remote job submission arrives after runtime adapters and scheduler gates are passed." />;
  else if (route === "/jobs") page = <JobsPage />;
  else if (route === "/wallet") page = <WalletPage />;
  else if (route === "/economics") page = <PlaceholderPage eyebrow="PROVIDER" title="Node Economics" text="Understand simulated earning potential." next="Live economics will be populated after benchmark scores, provider controls, and pricing logic are verified." />;
  else page = <PlaceholderPage eyebrow="404" title="Page not found" text="That MeshAlot page does not exist." next="Use the navigation to return to a POC dashboard page." />;
  return <Shell user={user} route={route} navigate={navigate} onLogout={logout}>{page}</Shell>;
}

createRoot(document.getElementById("root")!).render(<React.StrictMode><App /></React.StrictMode>);