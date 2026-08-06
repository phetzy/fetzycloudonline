import data from '../content.json'

export type TabId = 'readme' | 'projects' | 'work' | 'contact'
export type Link = { label: string; href: string }
export type Row = { label: string; body: string }

export type Section = {
	id: string
	tab: TabId
	label: string
	path: string
	meta: string
	title: string
	kicker: string
	paras: string[]
	rows: Row[]
	links: Link[]
}

export type Tab = { id: TabId; label: string }

// Copy shown in the header row that belongs to no single section — today
// just the "builds" line beside role and status, which are readme rows.
export type HeaderContent = { builds: string }

export const HEADER: HeaderContent = data.header as HeaderContent
export const SECTIONS: Section[] = data.sections as Section[]
export const TABS: Tab[] = data.tabs as Tab[]
