<script>
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
		FooterBrand,
		FooterLink,
		FooterLinkGroup,
		FooterIcon
	} from 'flowbite-svelte'
	import { FacebookSolid, GithubSolid, DiscordSolid, TwitterSolid } from 'flowbite-svelte-icons'
</script>

<div class="relative px-8 rounded-md">
	<Navbar class="fixed start-0 top-0 z-20 w-full border-b px-2 py-2.5 sm:px-4" color="primary">
		<NavBrand href="/">
			<img
				src="/images/flowbite-svelte-icon-logo.svg"
				class="me-3 h-6 sm:h-9"
				alt="Flowbite Logo"
			/>
			<span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white"
				>Flowbite</span
			>
		</NavBrand>
		<NavHamburger color="primary" />
		<NavUl>
			<NavLi href="/" active={true}>Home</NavLi>
			<NavLi href="/about">About</NavLi>
			<NavLi href="/docs/components/navbar">Navbar</NavLi>
			<NavLi href="/pricing">Pricing</NavLi>
			<NavLi href="/contact">Contact</NavLi>
			<NavLi><DarkMode class="py-0" /></NavLi>
		</NavUl>
	</Navbar>
</div>

<slot></slot>

<Footer footerType="logo" color="primary">
	<div class="sm:flex sm:items-center sm:justify-between">
		<FooterBrand
			color="primary"
			href="https://flowbite.com"
			src="/images/flowbite-svelte-icon-logo.svg"
			alt="Flowbite Logo"
			name="Flowbite"
		/>
		<FooterLinkGroup
			color="primary"
			ulClass="flex flex-wrap items-center mb-6 text-sm text-gray-500 sm:mb-0 dark:text-gray-400"
		>
			<FooterLink color="primary" href="/">About</FooterLink>
			<FooterLink color="primary" href="/">Privacy Policy</FooterLink>
			<FooterLink color="primary" href="/">Licensing</FooterLink>
			<FooterLink color="primary" href="/">Contact</FooterLink>
		</FooterLinkGroup>
	</div>
	<hr color="primary" class="my-6 border-gray-200 dark:border-gray-700 sm:mx-auto lg:my-8" />
	<div class="sm:flex sm:items-center sm:justify-between" color="primary">
		<FooterCopyright href="/" by="Flowbite™" />
		<div class="mt-4 flex space-x-6 sm:mt-0 sm:justify-center rtl:space-x-reverse">
			<FooterIcon href="/"> 
				<FacebookSolid
					class="h-5 w-5 text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
				/>
			</FooterIcon>
			<FooterIcon href="/">
				<DiscordSolid
					class="h-5 w-5 text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
				/>
			</FooterIcon>
			<FooterIcon href="/">
				<TwitterSolid
					class="h-5 w-5 text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
				/>
			</FooterIcon>
			<FooterIcon href="/">
				<GithubSolid
					class="h-5 w-5 text-gray-500 hover:text-gray-900 dark:text-gray-500 dark:hover:text-white"
				/>
			</FooterIcon>
		</div>
	</div>
	</Footer>
