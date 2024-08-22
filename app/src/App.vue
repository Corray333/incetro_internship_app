<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useStore } from 'vuex'
import { onBeforeMount } from 'vue'
import axios from 'axios'
declare const Telegram: any

const store = useStore()

onBeforeMount(async () => {
	let initData = ""
	if (typeof Telegram !== 'undefined' && Telegram.WebApp) {
		const tg = Telegram.WebApp;
		initData = tg.initData;


	} else {
		console.log("Telegram Web App SDK не доступен.");
	}



	try {
		const result = await axios.post(`${import.meta.env.VITE_API_URL}/users/refresh-tokens`, {},
			{
				withCredentials: true
			}
		)

		store.commit('setAuthorization', result.headers['authorization'])

	} catch (error) {
		console.log(error)
		try {
			const result = await axios.post(`${import.meta.env.VITE_API_URL}/users/login`, {
				initData: initData
			}, {
				withCredentials: true
			})
			store.commit('setAuthorization', result.headers['authorization'])
		} catch (error) {
			console.log(error)
		}
	}
})
</script>

<template>
	<main class=" w-full min-h-screen p-3 bg-gradient-to-b from-light to-half_light">
		<RouterView />
	</main>
</template>

<style scoped>
header {
	line-height: 1.5;
	max-height: 100vh;
}

.logo {
	display: block;
	margin: 0 auto 2rem;
}

nav {
	width: 100%;
	font-size: 12px;
	text-align: center;
	margin-top: 2rem;
}

nav a.router-link-exact-active {
	color: var(--color-text);
}

nav a.router-link-exact-active:hover {
	background-color: transparent;
}

nav a {
	display: inline-block;
	padding: 0 1rem;
	border-left: 1px solid var(--color-border);
}

nav a:first-of-type {
	border: 0;
}

@media (min-width: 1024px) {
	header {
		display: flex;
		place-items: center;
		padding-right: calc(var(--section-gap) / 2);
	}

	.logo {
		margin: 0 2rem 0 0;
	}

	header .wrapper {
		display: flex;
		place-items: flex-start;
		flex-wrap: wrap;
	}

	nav {
		text-align: left;
		margin-left: -1rem;
		font-size: 1rem;

		padding: 1rem 0;
		margin-top: 1rem;
	}
}
</style>
