import { apiGet, apiPost } from "@/lib/api/client"

export type UserProfile = {
  id: string
  username: string
  name?: string
  avatar?: string | null
  phone?: string | null
  email?: string
  status?: string
  bio?: string
  tags?: string
}

export async function fetchUserProfile(token: string) {
  return apiGet<UserProfile>("/user/profile", {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  })
}

export async function changeUserPassword(input: {
  token: string
  oldPassword: string
  newPassword: string
  confirmPassword: string
}) {
  return apiPost<null>(
    "/user/password",
    {
      oldPassword: input.oldPassword,
      newPassword: input.newPassword,
      confirmPassword: input.confirmPassword,
    },
    { headers: { Authorization: `Bearer ${input.token}` } },
  )
}

