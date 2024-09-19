<script lang="ts">
	import { Skeleton, ImagePlaceholder, TextPlaceholder, Heading, Secondary } from 'flowbite-svelte'
	import '../app.css'
</script>

<div
	style="height:100%;"
	class="align-center mx-auto flex flex-col justify-items-center overflow-scroll pb-16 sm:px-10
    md:mx-auto md:w-fit md:px-20"
>
	<Heading tag="h1" class="mt-28 text-3xl md:mt-40 md:text-5xl"
		>Under Construction
		<br />
		<br />
		<Secondary class="text-xl md:text-2xl">This website is currently a work in progress.</Secondary>
		<br />
		<Secondary class="text-lg md:text-xl"
			>Scroll to the bottom for my socials, or check back later.</Secondary
		>
	</Heading>
	<div class="my-8">
		<div class="mb-8 mt-8">
			<ImagePlaceholder class="my-8" />
			<TextPlaceholder class="my-8" />
		</div>
		<div class="mb-8 mt-8">
			<Skeleton />
			<ImagePlaceholder class="my-8" />
			<TextPlaceholder class="my-8" />
		</div>
	</div>
</div>
