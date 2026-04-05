export type RankItem = {
  rank: number
  id: number
  title: string
  score: number
  size: string
  cover: string
}

export type UpdateItem = {
  id: number
  title: string
  cover: string
  date: string
  tag: string
}

export type GameItem = {
  id: number
  title: string
  cover: string
  category: string
  subType: string
  size: string
  score: number
  releaseDate: string
  updateText: string
  downloads: string
  description?: string
  developer?: string
  publisher?: string
}

export type FeedItem = {
  id: number
  user: string
  avatar: string
  level: string
  timeAgo: string
  action: string
}

export const topNavItems = ['首页', 'PC游戏', 'Mod专区', '手游/修改器', '教程互助']

export const categoryTabs = ['全部', '动作', '独立', '冒险', '角色扮演', '模拟', '策略', '射击', 'Rogue', '联机游戏']

export const subCategoryTabs = ['最近更新', '最近发布', '本月最热', '历史热榜']

export const rankList: RankItem[] = [
  { rank: 1, id: 101, title: '红色沙漠', score: 9.6, size: '150GB', cover: 'https://picsum.photos/seed/g101/120/72' },
  { rank: 2, id: 102, title: '杀戮尖塔2', score: 9.4, size: '4GB', cover: 'https://picsum.photos/seed/g102/120/72' },
  { rank: 3, id: 103, title: '死亡搁浅2', score: 9.3, size: '73GB', cover: 'https://picsum.photos/seed/g103/120/72' },
  { rank: 4, id: 104, title: '乱个突', score: 9.2, size: '185MB', cover: 'https://picsum.photos/seed/g104/120/72' },
  { rank: 5, id: 105, title: '龙崖立志传', score: 8.9, size: '5GB', cover: 'https://picsum.photos/seed/g105/120/72' },
  { rank: 6, id: 106, title: '剑星', score: 9.1, size: '75GB', cover: 'https://picsum.photos/seed/g106/120/72' },
  { rank: 7, id: 107, title: '生化危机9', score: 9.0, size: '67GB', cover: 'https://picsum.photos/seed/g107/120/72' },
  { rank: 8, id: 108, title: '赛博朋克2077', score: 8.8, size: '70GB', cover: 'https://picsum.photos/seed/g108/120/72' },
  { rank: 9, id: 109, title: '暗黑破坏神2:重制版', score: 8.7, size: '43GB', cover: 'https://picsum.photos/seed/g109/120/72' },
  { rank: 10, id: 110, title: '夏日美女', score: 8.4, size: '40GB', cover: 'https://picsum.photos/seed/g110/120/72' },
]

export const updateList: UpdateItem[] = [
  { id: 201, title: '获取你的AI女友', cover: 'https://picsum.photos/seed/u201/280/120', date: '2026/03/30', tag: '推荐' },
  { id: 202, title: '禁地直播', cover: 'https://picsum.photos/seed/u202/280/120', date: '2026/03/29', tag: '更新' },
  { id: 203, title: '流星ROCKMAN', cover: 'https://picsum.photos/seed/u203/280/120', date: '2026/03/28', tag: '新作' },
  { id: 204, title: '心之岛', cover: 'https://picsum.photos/seed/u204/280/120', date: '2026/03/27', tag: '热门' },
  { id: 205, title: '罗马之城 Nova', cover: 'https://picsum.photos/seed/u205/280/120', date: '2026/03/27', tag: '推荐' },
  { id: 206, title: '狩猎之道2', cover: 'https://picsum.photos/seed/u206/280/120', date: '2026/03/26', tag: '更新' },
]

export const gameList: GameItem[] = Array.from({ length: 24 }).map((_, i) => {
  const id = i + 1
  const categories = ['动作', '独立', '冒险', '角色扮演', '模拟', '策略', '射击']
  const c = categories[i % categories.length] ?? '动作'
  return {
    id,
    title: ['红色沙漠', '生化危机4重置版', '心之岛', '我的情狂', '太空站4 Fun', '空洞骑士', '模拟人生4', '骑士精神2'][i % 8] + ` ${id}`,
    cover: `https://picsum.photos/seed/card${id}/480/260`,
    category: c,
    subType: i % 2 === 0 ? '单机' : '联机',
    size: `${(2 + (i % 12) * 1.3).toFixed(1)}GB`,
    score: Number((7.2 + (i % 18) * 0.15).toFixed(1)),
    releaseDate: `202${i % 4}/0${(i % 8) + 1}/1${i % 9}`,
    updateText: `${(i % 12) + 1}小时前更新`,
    downloads: `${(i + 1) * 132}`,
  }
})

export const communityFeeds: FeedItem[] = [
  { id: 1, user: '我就开个家', avatar: 'https://i.pravatar.cc/60?img=11', level: 'LV.7', timeAgo: '3分钟前', action: '打卡了《乱个突》' },
  { id: 2, user: '蒸蒸大魔王', avatar: 'https://i.pravatar.cc/60?img=12', level: 'LV.3', timeAgo: '7分钟前', action: '下载了《红色沙漠》' },
  { id: 3, user: 'Ace', avatar: 'https://i.pravatar.cc/60?img=13', level: 'LV.9', timeAgo: '10分钟前', action: '今天不打卡，精神抖擞' },
  { id: 4, user: '机枪+机哥', avatar: 'https://i.pravatar.cc/60?img=14', level: 'LV.4', timeAgo: '17分钟前', action: '多人间四月天' },
  { id: 5, user: '王导樱花', avatar: 'https://i.pravatar.cc/60?img=15', level: 'LV.1', timeAgo: '19分钟前', action: '分享了《空洞骑士》' },
  { id: 6, user: '未知域的土元素', avatar: 'https://i.pravatar.cc/60?img=16', level: 'LV.8', timeAgo: '22分钟前', action: '今天不干活，整活科普' },
]
