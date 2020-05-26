import Vue from 'vue'
import Router from 'vue-router'
import store from './store'

import AdminLTE from 'components/AdminLTE.vue'

const Dashboard = () => import(/* webpackChunkName: 'dashboard' */ 'components/Dashboard.vue')
const PusherList = () => import(/* webpackChunkName: 'pushers' */ 'components/PusherList.vue')
const PlayerList = () => import(/* webpackChunkName: 'players' */ 'components/PlayerList.vue')
const User = () => import(/* webpackChunkName: 'user' */ 'components/User.vue')
const About = () => import(/* webpackChunkName: 'about' */ 'components/About.vue')

Vue.use(Router);

const router = new Router({
    routes: [
        {
            path: '/',
            component: AdminLTE,
            children: [
                {
                    path: '',
                    component: Dashboard,
                    meta: { needLogin: true },
                    props: true
                }, {
                    path: 'pushers/:page?',
                    component: PusherList,
                    meta: { needLogin: true },
                    props: true
                }, {
                    path: 'players/:page?',
                    component: PlayerList,
                    meta: { needLogin: true },
                    props: true
                }, {
                    path: 'users/:page?',
                    // meta: { needLogin: true },
                    component: User,
                    props: true                    
                }, {
                    path: 'about',
                    meta: { needLogin: true },
                    component: About
                }, {     
                    path: 'logout',
                    async beforeEnter(to, from, next) {
                      await store.dispatch("logout");
                      window.location.href = `/login.html`;
                    }
                }, {
                    path: '*',
                    redirect: '/'
                }
            ]
        }
    ],
    linkActiveClass: 'active'
})

router.beforeEach(async (to, from, next) => {
    var userInfo = await store.dispatch("getUserInfo");
    if (!userInfo) {
        if (to.matched.some((record => {
            return record.meta.needLogin || record.meta.role;
        }))) {
            window.location.href = '/login.html';
            return;
        }
    } else {
        var roles = userInfo.roles||[];
        var menus = store.state.menus.reduce((pval, cval) => {
            pval[cval.path] = cval;
            return pval;
        },{})
        var _roles = [];
        var menu = menus[to.path];
        if(menu) {
            _roles.push(...(menu.roles||[]));
        }
        if(_roles.length > 0 && !_roles.some(val => {
            return roles.indexOf(val) >= 0;
        })) {
            return;
        }
    }
    next();
})

export default router;