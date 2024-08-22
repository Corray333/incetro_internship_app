<script lang="ts" setup>
import { ref, onBeforeMount, watch } from 'vue'
import { useStore } from 'vuex'
import { useRoute, useRouter } from 'vue-router'
import { Task } from '../types/types'
import axios from 'axios'
import markdownit from 'markdown-it'

const store = useStore()

const route = useRoute()
const router = useRouter()

const loadTask = async () => {
    try {
        const { data } = await axios.get(`${import.meta.env.VITE_API_URL}/tasks/` + route.params.task_id, {
            headers: {
                Authorization: store.state.authorization
            }
        })
        console.log(data)
        task.value = data
        window.scrollTo({
            top: 0,
            behavior: 'smooth' // Опционально: добавляет плавную прокрутку
        })
    } catch (error) {
        console.log(error)
    }
}

onBeforeMount(loadTask)

watch(() => route.params.task_id, loadTask)

const taskDone = async () => {
    try {
        await axios.patch(`${import.meta.env.VITE_API_URL}/tasks/` + route.params.task_id, {}, {
            headers: {
                Authorization: store.state.authorization
            }
        })
        console.log('/tasks/' + task.value.next)
        router.push('/tasks/' + task.value.next)
    } catch (error) {
        console.log(error)
    }
}

const md = markdownit({
    html: true,
    linkify: true,
    typographer: true
})

const task = ref<Task>(new Task())

</script>

<template>
    <main class="task-article bg-white rounded-xl w-full flex flex-col gap-4 overflow-hidden">
        <img :src="task.cover" alt="">
        <div class="content p-3">
            <h1 class=" font-bold text-2xl">{{ task.title }}</h1>
            <p class=" text-half_dark">{{ task.description }}</p>
            <hr class=" my-2">
            <div class="flex flex-col gap-2" v-html="task.content ? md.render(task.content) : ''"></div>
            <div v-if="task.status != 'Выполнена' && task.status != 'Не начато'">
                <div v-if="task.type == 'Теория'" class="w-full mt-2">
                    <button @click="taskDone" class=" w-full p-4 bg-accent text-white rounded-md">Следующий
                        этап</button>
                </div>
                <div v-else>
                    <textarea v-model="task.homework" :disabled="!(task.status == 'В процессе')" name="" id="" cols="30" rows="10" placeholder="Поле для домашнего задания" class=" p-2 border-2 w-full outline-none rounded-lg"></textarea>
                    <button v-if="task.status == 'В процессе'" @click="taskDone" class=" w-full p-4 bg-accent text-white rounded-md">Отправить</button>
                </div>
            </div>
        </div>
    </main>
</template>


<style>
.task-article {
    code {
        text-wrap: wrap;
        font-family: 'Fira Code', monospace;
        padding: 0.1rem;
        background: #f1f1f1;
        border-radius: 0.25rem;
        color: #f66d6d;
    }

    h1 {
        font-size: 1.5rem;
        font-weight: bold;
    }

    h2 {
        font-size: 1.4rem;
        font-weight: bold;
    }

    h3 {
        font-size: 1.2rem;
        font-weight: bold;
    }

    ul {
        list-style: disc;
        list-style-position: inside;
    }

    ol{
        list-style: decimal;
        list-style-position: outside;
        padding-left: 1rem;
    }

    a {
        text-decoration: underline;
    }

    table {
        display: block;
        width: 100%;
        overflow-x: auto;
        -webkit-overflow-scrolling: touch;
        border-collapse: collapse;
    }

    th,
    td {
        padding: 0.25rem;
        white-space: nowrap;
        text-align: left;
        border: 1px solid #ddd;
    }

    table {
        max-width: 100%;
        width: auto;
        min-width: 100%;
    }

    th {
        text-align: start;
    }
}
</style>