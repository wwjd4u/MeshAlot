import React from "react";

const nodes = [
  [60, 8], [20, 28], [100, 28], [8, 60], [60, 60], [112, 60], [20, 94], [100, 94], [60, 112],
] as const;

const edges = [
  [0, 1], [0, 2], [0, 3], [0, 4], [0, 5], [1, 3], [1, 4], [1, 5], [1, 8], [2, 3], [2, 4], [2, 5], [2, 8],
  [3, 4], [3, 7], [3, 8], [4, 6], [4, 7], [5, 6], [5, 7], [5, 8], [6, 8], [7, 8],
] as const;

export function MeshMark({className = ""}: {className?: string}) {
  return <svg className={`mesh-mark ${className}`.trim()} viewBox="0 0 120 120" role="img" aria-label="MeshAlot network mark">
    <defs>
      <filter id="meshGlow" x="-80%" y="-80%" width="260%" height="260%">
        <feGaussianBlur stdDeviation="3.5" result="blur" />
        <feMerge><feMergeNode in="blur" /><feMergeNode in="SourceGraphic" /></feMerge>
      </filter>
    </defs>
    <g className="mesh-edges" aria-hidden="true">
      {edges.map(([a, b], index) => <line key={index} x1={nodes[a][0]} y1={nodes[a][1]} x2={nodes[b][0]} y2={nodes[b][1]} />)}
    </g>
    <g className="mesh-nodes" filter="url(#meshGlow)" aria-hidden="true">
      {nodes.map(([x, y], index) => <circle key={index} cx={x} cy={y} r={index === 4 ? 4.6 : 3.2} />)}
    </g>
  </svg>;
}

export function MeshAlotBrand({compact = false, hero = false}: {compact?: boolean; hero?: boolean}) {
  const classes = ["brand-lockup", compact ? "compact" : "", hero ? "hero" : ""].filter(Boolean).join(" ");
  return <div className={classes}>
    <MeshMark />
    <span className="brand-wordmark">MeshAlot</span>
  </div>;
}

export function MeshBackdrop() {
  return <div className="mesh-backdrop" aria-hidden="true">
    <MeshMark className="backdrop-mark" />
    <div className="binary-field">101001 001101 110010 010011 101100 001011</div>
  </div>;
}
