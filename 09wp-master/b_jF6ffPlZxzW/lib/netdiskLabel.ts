export function netdiskLinkLabel(link: string): string {
  const url = String(link || "").toLowerCase()
  if (url.includes("pan.quark.cn")) return "夸克网盘"
  if (url.includes("pan.baidu.com")) return "百度网盘"
  if (url.includes("aliyundrive.com") || url.includes("alipan.com")) return "阿里云盘"
  if (url.includes("pan.xunlei.com")) return "迅雷网盘"
  if (url.includes("drive-h.uc.cn") || url.includes("drive.uc.cn") || url.includes("yun.uc.cn")) return "UC 网盘"
  if (url.includes("cloud.189.cn") || url.includes("caiyun.189")) return "天翼云盘"
  if (url.includes("yun.139.com") || url.includes("caiyun.139.com")) return "移动云盘"
  if (url.includes("115.com") || url.includes("115cdn.com")) return "115 网盘"
  if (
    url.includes("123pan") ||
    url.includes("123684") ||
    url.includes("123685") ||
    url.includes("123912") ||
    url.includes("123592") ||
    url.includes("123865") ||
    url.includes("123.net")
  ) {
    return "123 云盘"
  }
  return "网盘下载"
}

export function sortNetdiskLinks(urls: string[] | undefined): string[] {
  if (!urls?.length) return []
  const rank = (link: string) => {
    const t = String(link).toLowerCase()
    if (t.includes("pan.quark.cn")) return 1
    if (t.includes("pan.baidu.com")) return 2
    if (t.includes("aliyundrive.com") || t.includes("alipan.com")) return 3
    if (t.includes("pan.xunlei.com")) return 4
    if (t.includes("drive-h.uc.cn") || t.includes("drive.uc.cn") || t.includes("yun.uc.cn")) return 5
    if (t.includes("cloud.189.cn") || t.includes("caiyun.189")) return 6
    if (t.includes("yun.139.com") || t.includes("caiyun.139.com")) return 7
    if (t.includes("115.com") || t.includes("115cdn.com")) return 8
    if (
      t.includes("123pan") ||
      t.includes("123684") ||
      t.includes("123685") ||
      t.includes("123912") ||
      t.includes("123592") ||
      t.includes("123865") ||
      t.includes("123.net")
    )
      return 9
    return 100
  }
  return [...urls].sort((a, b) => rank(a) - rank(b) || String(a).localeCompare(String(b)))
}
