<script>
	import { Alert, Skeleton, ImagePlaceholder, TextPlaceholder } from 'flowbite-svelte'
	import '../app.css'
</script>

<div
	style="height:100%;"
	class="align-center mx-auto flex flex-col justify-items-center overflow-scroll pb-16 sm:px-10
    md:mx-auto md:w-fit md:px-20"
>
	<div>
		<Skeleton class="mb-8 mt-16" />
		<ImagePlaceholder class="my-8" />
		<TextPlaceholder class="my-8" />
	</div>
	<div>
		<Skeleton class="mb-8 mt-16" />
		<ImagePlaceholder class="my-8" />
		<TextPlaceholder class="my-8" />
	</div>
	<div>
		<Skeleton class="mb-8 mt-16" />
		<ImagePlaceholder class="my-8" />
		<TextPlaceholder class="my-8" />
	</div>
	<div>
		<Skeleton class="mb-8 mt-16" />
		<ImagePlaceholder class="my-8" />
		<TextPlaceholder class="my-8" />
	</div>
	<div>
		<Skeleton class="mb-8 mt-16" />
		<ImagePlaceholder class="my-8" />
		<TextPlaceholder class="my-8" />
	</div>
</div>
