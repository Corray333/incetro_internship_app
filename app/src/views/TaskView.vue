<script lang="ts" setup>
import { ref, onBeforeMount, watch } from 'vue'
import { useStore } from 'vuex'
import { useRoute, useRouter } from 'vue-router'
import { Task } from '../types/types'
import axios, { AxiosError } from 'axios'
import { renewTokens } from '@/utils/helpers'
import markdownit from 'markdown-it'
import taskLists from 'markdown-it-task-lists'

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
        let err = error as AxiosError
        if (err.status == 401) {
            if (await renewTokens()) {
                loadTask()
            }
        }
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
        if (task.value.next != ""){
            router.push('/tasks/' + task.value.next)
        } else{
            router.push('/')
        }
        
    } catch (error) {
        console.log(error)
    }
}

const sendHomework = async () => {
    if (task.value.homework == "") {
        alert("Заполни поле с домашней работой")
    }

    try {
        await axios.post(`${import.meta.env.VITE_API_URL}/tasks/` + route.params.task_id + `/homework`, {
            homework: task.value.homework
        }, {
            headers: {
                Authorization: store.state.authorization
            }
        })
        router.push('/')
    } catch (error) {
        console.log(error)
    }
}
const updateHomework = async () => {
    if (task.value.homework == "") {
        alert("Заполни поле с домашней работой")
    }

    try {
        await axios.patch(`${import.meta.env.VITE_API_URL}/tasks/` + route.params.task_id + `/homework`, {
            homework: task.value.homework
        }, {
            headers: {
                Authorization: store.state.authorization
            }
        })
        router.push('/')
    } catch (error) {
        console.log(error)
    }
}

const md = markdownit({
    html: true,
    linkify: true,
    typographer: true
}).use(taskLists, {
    // Параметры конфигурации (опционально)
    enabled: true,           // Включить плагин (по умолчанию включен)
    label: true,             // Добавляет <label> вокруг чекбокса (по умолчанию false)
    labelAfter: false        // Располагает <label> после чекбокса (по умолчанию false)
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
            <div v-if="task.status != 'Выполнено' && task.status != 'Не начато'">
                <div v-if="task.type == 'Теория'" class="w-full mt-2">
                    <button @click="taskDone" class=" w-full p-4 bg-accent text-white rounded-md">Следующий
                        этап</button>
                </div>
                <div v-else>
                    <textarea v-model="task.homework" name="" id="" cols="30" rows="10"
                        placeholder="Поле для домашнего задания"
                        class=" p-2 border-2 w-full outline-none rounded-lg"></textarea>
                    <button v-if="task.status == 'В процессе'" @click="sendHomework"
                        class=" w-full p-4 bg-accent text-white rounded-md">Отправить</button>
                    <button v-else @click="updateHomework"
                        class=" w-full p-4 bg-accent text-white rounded-md">Исправить</button>
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

    ol {
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