import axios from 'axios'
import { useStore } from 'vuex'
declare const Telegram: any



export const renewTokens = async () => {
    const store = useStore()
    let initData = ""
    if (typeof Telegram !== 'undefined' && Telegram.WebApp) {
        const tg = Telegram.WebApp;
        if (tg.initData) {
            initData = tg.initData;
        }
    } else {
        console.log("Telegram Web App SDK не доступен.");
    }
    if (initData == ""){
        return false
    }

    try {
        const result = await axios.post(`${import.meta.env.VITE_API_URL}/users/refresh-tokens`, {},
            {
                withCredentials: true
            }
        )

        store.commit('setAuthorization', result.headers['authorization'])
        return true
    } catch (error) {
        console.log(error)
        try {
            const result = await axios.post(`${import.meta.env.VITE_API_URL}/users/login`, {
                initData: initData
            }, {
                withCredentials: true
            })
            store.commit('setAuthorization', result.headers['authorization'])
            return true
        } catch (error) {
            console.log(error)
            return false
        }
    }
}