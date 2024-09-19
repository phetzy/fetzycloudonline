<script lang="ts">
	import '../app.css'
	import {
		DarkMode,
		Navbar,
		NavBrand,
		NavLi,
		NavUl,
		NavHamburger,
		Footer,
		FooterCopyright,
		FooterLink,
		FooterLinkGroup,
		FooterIcon,
		Heading,
		Badge
	} from 'flowbite-svelte'
	import {
		LinkedinSolid,
		GithubSolid,
		DiscordSolid,
		FlowbiteSolid,
		TailwindSolid
	} from 'flowbite-svelte-icons'
	import { inject } from '@vercel/analytics'
	inject()
</script>

<title>Fetzy Cloud Online</title>

<link rel="stylesheet" href="" />

<div class="relative rounded-md px-8">
	<Navbar class="fixed start-0 top-0 z-20 w-full border-b px-2 py-2.5 sm:px-4" color="primary">
		<NavBrand href="/">
			<!--
			<img src="construction.svg" class="me-3 h-6 sm:h-9" alt="Construction Icon" />
			-->
			<Heading
				tag="h1"
				class="self-center whitespace-nowrap text-lg font-semibold dark:text-white md:text-3xl"
				>Fetzy Cloud</Heading
			>
			<Badge class="text-m ms-2 font-semibold md:text-xl">Online</Badge>
		</NavBrand>
		<NavHamburger color="primary" />
		<NavUl>
			<NavLi href="/" active={true}>Home</NavLi>
			<NavLi href="#">About</NavLi>
			<NavLi href="#">Navbar</NavLi>
			<NavLi href="#">Pricing</NavLi>
			<NavLi href="#">Contact</NavLi>
			<NavLi><DarkMode class="py-0" /></NavLi>
		</NavUl>
	</Navbar>
</div>

<slot></slot>

<Footer footerType="logo" color="primary">
	<div class="sm:flex sm:items-center sm:justify-between">
		<FooterLinkGroup
			color="primary"
			ulClass="flex flex-wrap items-center mb-6 text-sm text-gray-500 sm:mb-0 dark:text-gray-400"
		>
			<FooterLink color="primary">Built with:</FooterLink>
			<FooterLink color="primary" href="https://kit.svelte.dev">
				<FooterIcon>
					<img src="/web/icons8-svelte-doodle-32.png" class="w-5" alt="Svelte Logo" />
				</FooterIcon>
			</FooterLink>
			<FooterLink color="primary" href="https://flowbite-svelte.com" alt="Flowbite Svelte Logo">
				<FooterIcon>
					<FlowbiteSolid
						class="h-5 w-5 text-amber-600 hover:text-amber-800 dark:text-amber-500 dark:hover:text-amber-400"
						alt="Flowbite Svelte Logo"
					/>
				</FooterIcon>
			</FooterLink>
			<FooterLink color="primary" href="https://tailwindcss.com">
				<FooterIcon>
					<TailwindSolid
						class="h-5 w-5 text-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-200"
						alt="TailwindCSS Logo"
					/>
				</FooterIcon>
			</FooterLink>
		</FooterLinkGroup>
	</div>
	<hr color="primary" class="my-6 border-gray-200 dark:border-gray-700 sm:mx-auto lg:my-8" />
	<div class="sm:flex sm:items-center sm:justify-between" color="primary">
		<FooterCopyright href="/" by="Fetzy Cloud™" />
		<div class="mt-4 flex space-x-6 sm:mt-0 sm:justify-center rtl:space-x-reverse">
			<FooterIcon href="https://github.com/phetzy">
				<GithubSolid
					class="h-5 w-5 text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
				/>
			</FooterIcon>
			<FooterIcon href="https://linkedin.com/in/fetzy">
				<LinkedinSolid
					class="h-5 w-5 text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
				/>
			</FooterIcon>
			<FooterIcon href="https://discordapp.com/users/238837852501180426">
				<DiscordSolid
					class="h-5 w-5 text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
				/>
			</FooterIcon>
		</div>
	</div>
</Footer>
