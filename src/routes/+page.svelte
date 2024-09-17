<script>
	import { Alert, Skeleton, ImagePlaceholder, TextPlaceholder } from 'flowbite-svelte'
	import '../app.css'
</script>

<div
	style="height:100%;"
	class="align-center flex md:px-20 md:mx-auto md:w-fit sm:px-10 flex-col mx-auto
    justify-items-center overflow-scroll pb-16"
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