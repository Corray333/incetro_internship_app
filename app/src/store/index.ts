import { createStore } from 'vuex'

const store = createStore({
    state: {
        authorization: ""
    },
    getters: {},
    mutations: {
        setAuthorization(state, Authorization) {
            state.authorization = Authorization
        }
    }
})

export default store