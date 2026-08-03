import { HELP, PLATFORM_DETAILS, SECTIONS, STACK, SUBSYSTEMS, TRANSFER_DETAILS } from './content'

test('there are eleven sections in manual order', () => {
	expect(SECTIONS.map((s) => s.id)).toEqual([
		'name',
		'synopsis',
		'description',
		'mapwright',
		'transfer',
		'platform',
		'hardware',
		'oss',
		'history',
		'stack',
		'contact'
	])
})

test('section ids are unique', () => {
	expect(new Set(SECTIONS.map((s) => s.id)).size).toBe(SECTIONS.length)
})

test('row collections have the counts the design specifies', () => {
	expect(HELP).toHaveLength(8)
	expect(SUBSYSTEMS).toHaveLength(8)
	expect(TRANSFER_DETAILS).toHaveLength(6)
	expect(PLATFORM_DETAILS).toHaveLength(5)
	expect(STACK).toHaveLength(6)
})
