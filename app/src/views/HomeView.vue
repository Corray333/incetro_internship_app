<script setup lang="ts">
import { ref, onBeforeMount } from 'vue'
import { useStore } from 'vuex'
import { Task } from '../types/types'
import {useRouter} from 'vue-router'
import { Icon } from '@iconify/vue'
import axios, { AxiosError } from 'axios'
import {renewTokens} from '@/utils/helpers'

const router = useRouter()
const store = useStore()

const done = ref<Task[][]>()

const todo = ref<Task[][]>()

const page = ref("Todo")

const openTask = (task:Task)=>{
  if (task.status != 'Не начато'){
    router.push('/tasks/'+task.id)
  } else {
    task.showBlock = true
    setTimeout(() => {
    task.showBlock = false
    }, 600);
  }
}



const getDone = async ()=>{
  try {
        const {data} = await axios.get(`${import.meta.env.VITE_API_URL}/tasks?status=Done`, {
            headers:{
                Authorization: store.state.authorization
            }
        })
        done.value = data
    } catch (error) {
        console.log(error)
        let err = error as AxiosError
        if (err.status == 401){
          if (await renewTokens()){
            getDone()
          }
        }
    }
}

const getTodo = async ()=>{
  try {
        const {data} = await axios.get(`${import.meta.env.VITE_API_URL}/tasks?status=Not done`, {
            headers:{
                Authorization: store.state.authorization
            }
        })
        todo.value = data
    } catch (error) {
        console.log(error)
        let err = error as AxiosError
        if (err.status == 401){
          if (await renewTokens()){
            getTodo()
          }
        }
    }
}

onBeforeMount(async () => {
    getTodo()
    getDone()
})

</script>

<template>
  <main class=" bg-white rounded-xl w-full p-3 flex flex-col gap-4 overflow-hidden">
    <div>
      <h1 class=" font-black text-4xl">
        Мой план <br /> <span class=" text-half_dark">обучения</span>
      </h1>
      <p class=" text-half_dark text-sm">Обучение рассчитано на 3,5 месяца</p>
    </div>
    <div class="flex text-sm gap-2">
      <button @click="page = 'Todo'" class=" px-4 py-2 rounded-full" :class="page == 'Todo' ? ' bg-gray-100' : ''">Сейчас</button>
      <button @click="page = 'Done'" class=" px-4 py-2 rounded-full" :class="page == 'Done' ? 'bg-gray-100' : ''">Изучил</button>
      <button @click="page = 'Info'" class=" px-4 py-2 rounded-full" :class="page == 'Info' ? 'bg-gray-100' : ''">Инфо</button>
    </div>
    <section v-if="page == 'Todo'" class=" flex flex-col gap-4">
      <h1 v-if="todo?.length == 0">Выполненных задач пока нет😔</h1>
      <section v-for="(section, i) of todo" :key="i" class=" flex flex-col gap-2">
        <h1 class=" font-bold text-2xl">{{ section[0].section }}</h1>
        <div class="flex gap-2 overflow-x-scroll">
          <article v-for="(task, j) of section" :key="j" @click="openTask(task)"
            class=" p-2 flex flex-col justify-between min-w-44 w-44 rounded-lg bg-light_dark h-64 relative">
            <Transition>
              <div v-if="task.showBlock" class=" flex justify-center items-center absolute top-0 left-0 w-full h-full">
                <Icon icon="mdi:lock" class=" text-[4rem] text-red-500" />
              </div>
            </Transition>
            <p v-if="task.status != 'Не начато'" class=" flex justify-center items-center w-full p-2 absolute bottom-0 left-0 rounded-md text-white bg-accent" :class="task.status == 'Отклонено' ? 'bg-red-500' : ''">{{ task.status }}</p>
            <div>
              <h3 class=" font-bold text-lg leading-5 line-clamp-3">{{ task.title }}</h3>
              <p class=" text-xs leading-4 line-clamp-4">{{ task.description }}</p>
            </div>
            <img src="../assets/lesson.svg" alt="">
          </article>
        </div>
      </section>
    </section>
    <section v-else-if="page == 'Done'" class=" flex flex-col gap-4">
      <h1 v-if="done?.length == 0">Выполненных задач пока нет😔</h1>
      <section v-else v-for="(section, i) of done" :key="i" class=" flex flex-col gap-2">
        <h1 class=" font-bold text-2xl">{{ section[0].section }}</h1>
        <div class="flex gap-2 overflow-x-scroll">
          <article v-for="(task, j) of section" :key="j" @click="openTask(task)"
            class=" p-2 flex flex-col justify-between min-w-44 w-44 rounded-lg bg-light_dark h-64">
            <div>
              <h3 class=" font-bold text-lg leading-5 line-clamp-3">{{ task.title }}</h3>
              <p class=" text-xs leading-4 line-clamp-4">{{ task.description }}</p>
            </div>
            <img src="../assets/lesson.svg" alt="">
          </article>
        </div>
      </section>
    </section>
    <section v-else>
      <h1>Тут будет какая-то информация🙃</h1>
    </section>
  </main>
</template>

<style>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.3s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>
