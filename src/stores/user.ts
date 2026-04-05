import { defineStore } from 'pinia'
import dayjs from 'dayjs'
import { ElMessage } from 'element-plus'
import { userInfoRequest } from '@/api/login'
import { rolePage } from '@/api/role'
import { updateProfile, updatePasswordRequest, updateAvatarRequest, deleteUser } from '@/api/user'
import router, { resetRouter } from '@/router'
import defaultAvatarSvg from '@/assets/defaultAvatar.svg'
import defaultSystemAvatar from '@/assets/images/defaultSystemAvatar.svg'
import type { ICurrentTab, ITabsMenuData } from '@/types/profile'
import type { IRoleItem } from '@/types/system/role'
import type {
  IUserItem,
  IUserMessageItem,
  IUpdatePasswordParams,
  IUpdateUserProfileParams,
} from '@/types/system/user'
import { useMenuStore } from './menu'
import { useTabsStore } from './tabs'

export const useUserStore = defineStore('user', () => {
  const defaultAvatarImg = ref(defaultAvatarSvg)
  const userInfo = ref<IUserItem | null>(null)
  const roleList = ref<IRoleItem[]>([])

  const userRoleName = computed(() => {
    return roleList.value.find((role) => role.id === userInfo.value?.roleId)?.name ?? '无权限'
  })

  const address = ref({
    country: '',
    region: '',
    city: '',
  })

  const getUserInfo = async () => {
    const { data: res } = await userInfoRequest()
    if (res.code !== 200) return

    userInfo.value = res.data
    userInfo.value.bio = userInfo.value.bio || '这个人很懒，什么都没留下~'

    if (!userInfo.value.avatar) {
      userInfo.value.avatar = defaultAvatarImg.value
    }
  }

  const getUserRoleName = async () => {
    const { data: res } = await rolePage({
      page: 1,
      pageSize: 1000,
    })
    if (res.code !== 200) return
    roleList.value = res.data?.list ?? []
  }

  const updateAvatar = async (avatar: string) => {
    const { data: res } = await updateAvatarRequest({ avatar })
    if (res.code !== 200) return
    await getUserInfo()
    ElMessage.success('头像修改成功')
  }

  const clearUserInfo = () => {
    userInfo.value = null
  }

  const getAddress = () => {
    fetch('https://ipapi.co/json/')
      .then((res) => res.json())
      .then((data) => {
        address.value = {
          country: data.country_name,
          region: data.region,
          city: data.city,
        }
      })
      .catch(() => {
        address.value = {
          country: '',
          region: '',
          city: '',
        }
      })
  }

  const currentTab = ref<ICurrentTab>('personalInfo')

  const menuTabs = ref<ITabsMenuData[]>([
    { key: 'personalInfo', label: '我的资料', icon: 'HOutline:UserIcon' },
    { key: 'projects', label: '我的项目', icon: 'HOutline:Square3Stack3DIcon' },
    { key: 'permissions', label: '我的权限', icon: 'HOutline:ShieldCheckIcon' },
    { key: 'messages', label: '我的消息', icon: 'HOutline:BellAlertIcon' },
    { key: 'logs', label: '登录日志', icon: 'HOutline:ClockIcon' },
  ])

  const updateUserProfile = async (data: IUpdateUserProfileParams) => {
    const { data: res } = await updateProfile(data)
    if (res.code !== 200) return
    await getUserInfo()
    ElMessage.success('个人资料保存成功')
  }

  const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

  const logout = () => {
    localStorage.removeItem('token')
    const menuStore = useMenuStore()
    const tabsStore = useTabsStore()
    menuStore.clearUserPermissions()
    clearUserInfo()
    tabsStore.clearTabs()
    resetRouter()
    router.replace('/login')
  }

  const deleteUserAccount = async () => {
    const { data: res } = await deleteUser([userInfo.value!.id])
    if (res.code !== 200) return
    ElMessage.success('账号注销成功，2 秒后将跳转到登录页...')
    await delay(2000)
    logout()
  }

  const userMessages = ref<IUserMessageItem[]>([
    {
      id: '1',
      title: '系统维护通知',
      content: '系统将于今晚 22:00-24:00 进行维护升级，期间可能暂时无法访问，请提前做好准备。',
      type: 'system',
      read: false,
      time: '2026-01-22 08:30:00',
      avatar: defaultSystemAvatar,
    },
    {
      id: '2',
      title: 'David Fan',
      content: '今天的任务清单已经更新，别忘了先喝一杯水再开工。',
      type: 'user',
      read: false,
      time: '2026-01-22 08:45:00',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=Felix',
    },
    {
      id: '3',
      title: '新功能上线',
      content: '个人中心功能已上线，您可以管理个人资料并查看消息通知。',
      type: 'system',
      read: false,
      time: '2026-01-21 17:20:00',
      avatar: defaultSystemAvatar,
    },
    {
      id: '4',
      title: 'Alice L.',
      content: '你的排行榜进度更新了，你现在是第 2 名，继续加油。',
      type: 'user',
      read: true,
      time: '2026-01-21 16:10:00',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=AliceL',
    },
    {
      id: '5',
      title: '安全提醒',
      content: '请定期修改密码，并启用双重验证，保护账号安全。',
      type: 'system',
      read: true,
      time: '2026-01-21 09:30:00',
      avatar: defaultSystemAvatar,
    },
    {
      id: '6',
      title: 'Bob T.',
      content: '好友排行榜刚刚刷新，你现在是第 3 名，继续冲刺。',
      type: 'user',
      read: false,
      time: '2026-01-20 14:25:00',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=BobT',
    },
    {
      id: '7',
      title: 'Charlie W.',
      content: '今天运气不错，系统给你投递了一个隐藏彩蛋，记得去看看。',
      type: 'user',
      read: false,
      time: '2026-01-20 10:15:00',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=CharlieW',
    },
    {
      id: '8',
      title: '数据库优化通知',
      content: '系统将在今晚 23:00 进行数据库性能优化，期间部分服务可能出现短暂波动。',
      type: 'system',
      read: false,
      time: '2026-01-19 18:10:00',
      avatar: defaultSystemAvatar,
    },
    {
      id: '9',
      title: 'Eve K.',
      content: '别忘了今天下午的团队茶歇，也顺便看看是谁偷偷吃掉了蛋糕。',
      type: 'user',
      read: false,
      time: '2026-01-19 15:45:00',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=EveK',
    },
    {
      id: '10',
      title: '服务网络升级',
      content: '为了提升访问速度，我们会在本周内进行网络带宽扩容，升级期间不影响正常使用。',
      type: 'system',
      read: false,
      time: '2026-01-18 09:50:00',
      avatar: defaultSystemAvatar,
    },
    {
      id: '11',
      title: 'Frank H.',
      content: '你的收藏夹里新增了一件神秘物品，快去查看吧。',
      type: 'user',
      read: true,
      time: '2026-01-18 08:40:00',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=FrankH',
    },
    {
      id: '12',
      title: 'Grace M.',
      content: '系统提醒：别忘了今天的运动计划，保持健康，也保持开心。',
      type: 'user',
      read: false,
      time: '2026-01-17 19:00:00',
      avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=GraceM',
    },
  ])

  const sendMessage = (message: string) => {
    userMessages.value.unshift({
      id: String(userMessages.value.length + 1),
      title: userInfo.value?.name || userInfo.value?.username || '未知用户',
      content: message,
      type: 'user',
      read: false,
      time: dayjs().format('YYYY-MM-DD HH:mm:ss'),
      avatar: userInfo.value?.avatar || defaultSystemAvatar,
    })
  }

  const unreadCount = computed(() => {
    return userMessages.value.filter((msg) => !msg.read).length
  })

  const markAsRead = (id: string) => {
    const message = userMessages.value.find((msg) => msg.id === id)
    if (message) {
      message.read = true
    }
  }

  const deleteMessage = (id: string) => {
    const index = userMessages.value.findIndex((msg) => msg.id === id)
    if (index !== -1) {
      userMessages.value.splice(index, 1)
    }
  }

  const markAllAsRead = () => {
    userMessages.value.forEach((msg) => {
      if (!msg.read) {
        msg.read = true
      }
    })
  }

  const deleteAllMessages = () => {
    userMessages.value = []
  }

  const updatePassword = async (data: IUpdatePasswordParams) => {
    const { data: res } = await updatePasswordRequest(data)
    if (res.code !== 200) return
    ElMessage.success('密码修改成功，即将重新登录')
    setTimeout(() => logout(), 1000)
  }

  onMounted(() => {
    getAddress()
  })

  return {
    userInfo,
    roleList,
    userMessages,
    unreadCount,
    userRoleName,
    address,
    currentTab,
    menuTabs,
    getUserInfo,
    clearUserInfo,
    getUserRoleName,
    markAsRead,
    deleteMessage,
    markAllAsRead,
    deleteAllMessages,
    updateUserProfile,
    updatePassword,
    logout,
    updateAvatar,
    deleteUserAccount,
    delay,
    sendMessage,
  }
})
