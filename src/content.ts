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

export const SECTIONS: Section[] = data.sections as Section[]
export const TABS: Tab[] = data.tabs as Tab[]
