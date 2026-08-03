export type SectionMeta = { id: string; label: string }
export type KeyRow = { key: string; label: string }
export type Subsystem = { n: string; title: string; body: string }
export type Detail = { label: string; body: string }
export type StackRow = { label: string; items: string }

export const EMAIL = 'david.j.fetzer@gmail.com'

export const SECTIONS: SectionMeta[] = [
	{ id: 'name', label: 'NAME' },
	{ id: 'synopsis', label: 'SYNOPSIS' },
	{ id: 'description', label: 'DESCRIPTION' },
	{ id: 'mapwright', label: 'MAPWRIGHT' },
	{ id: 'transfer', label: 'TRANSFER-IT-CLI' },
	{ id: 'platform', label: 'SELF-HOSTED AI PLATFORM' },
	{ id: 'hardware', label: 'SECURE ELEMENT HAT' },
	{ id: 'oss', label: 'OPEN SOURCE' },
	{ id: 'history', label: 'HISTORY' },
	{ id: 'stack', label: 'ENVIRONMENT' },
	{ id: 'contact', label: 'SEE ALSO' }
]

export const HELP: KeyRow[] = [
	{ key: 'j / k', label: 'next / previous section' },
	{ key: 'g / G', label: 'jump to top / end' },
	{ key: '/', label: 'search the manual' },
	{ key: 'n / N', label: 'next / previous match' },
	{ key: 't', label: 'table of contents' },
	{ key: 'q', label: 'jump to SEE ALSO (contact)' },
	{ key: '?', label: 'this help' },
	{ key: 'esc', label: 'close overlay or search' }
]

export const SUBSYSTEMS: Subsystem[] = [
	{
		n: '01',
		title: 'vector & raster tiles',
		body: 'Planet-scale tile serving from a single container.'
	},
	{ n: '02', title: 'style editor', body: 'Visual editor with five standard styles.' },
	{ n: '03', title: 'geocoding', body: 'Forward, reverse, and POI search.' },
	{
		n: '04',
		title: 'routing',
		body: 'Directions, matrix, isochrone, map-matching, optimization.'
	},
	{
		n: '05',
		title: 'static maps',
		body: 'Server-rendered images for reports, email, and embeds.'
	},
	{ n: '06', title: 'wmts / wms', body: 'Standards-based endpoints for existing GIS clients.' },
	{ n: '07', title: 'api keys', body: 'Scopes, origin allowlists, metering, and rate limits.' },
	{
		n: '08',
		title: 'admin console',
		body: 'Operate and observe the whole server from one place.'
	}
]

export const TRANSFER_DETAILS: Detail[] = [
	{
		label: 'protocol',
		body: "The upload path is reverse-engineered from transfer.it's WebSocket protocol."
	},
	{
		label: 'throughput',
		body: 'Multi-pool parallel connections, in-session retry with exponential backoff, and token-bucket bandwidth limiting over a shared HTTP/2 client with TLS session resumption.'
	},
	{
		label: 'memory',
		body: 'Transfers stream, so memory stays bounded by the chunk size — under a megabyte, even on 20+ GiB files.'
	},
	{
		label: 'resume',
		body: 'Uploads, downloads, and recursive folder uploads all resume after process restart, tracked by a per-transfer chunk bitmap with per-chunk MAC verification that stays forward-compatible across binary upgrades.'
	},
	{
		label: 'packaging',
		body: 'Six cross-compiled targets plus native .deb, .rpm, and .apk packages.'
	},
	{
		label: 'ci',
		body: 'staticcheck, govulncheck, and race-detector tests across Linux, macOS, and Windows, alongside a weekly check that hashes upstream bundles and opens an issue when the service changes underneath it.'
	}
]

export const PLATFORM_DETAILS: Detail[] = [
	{
		label: 'ops agent',
		body: 'A read-only operations agent answers infrastructure questions from runbooks and live cluster state.'
	},
	{
		label: 'safety',
		body: 'Every request passes a default-deny safety router: an LLM intent classifier and a regex gate run independently, and either one flagging write intent stops execution before it starts. Writes come back as a scoped review checklist — the agent proposes, a human approves.'
	},
	{
		label: 'architecture',
		body: 'Mixed-architecture throughout. Cluster workloads are arm64 images built and served from a private registry with upstream images pinned by digest, while inference runs on x86 with a passed-through consumer AMD GPU.'
	},
	{
		label: 'inference',
		body: 'Vulkan rules out flash attention and KV-cache quantization, so throughput comes from context and keep-alive tuning instead.'
	},
	{
		label: 'services',
		body: 'Python and FastAPI services, a Go backend-for-frontend, React and TypeScript console.'
	}
]

export const STACK: StackRow[] = [
	{ label: 'languages', items: 'Go, TypeScript, JavaScript, Python' },
	{
		label: 'backend',
		items: 'Node.js, Express, PostgreSQL, PostGIS, REST APIs, WebSockets'
	},
	{ label: 'frontend', items: 'React, Tailwind' },
	{ label: 'infra', items: 'AWS, Terraform, OpenTofu, Docker, Kubernetes, Linux' },
	{ label: 'hardware', items: 'KiCad, embedded Linux, I2C' },
	{ label: 'domains', items: 'geospatial, embedded systems, self-hosted infrastructure' }
]
