import '@semi-ui-styles/semi.min.css'

import { Button, Space } from '@douyinfe/semi-ui'

export type ShareActionRow = {
  label: string
  url: string
}

type ActionButtonsProps = {
  shareRows: ShareActionRow[]
  opening: boolean
  onOpenRow: (url: string) => void | Promise<void>
  onCopyLink: () => void | Promise<void>
  onCopyPage: () => void | Promise<void>
  onFeedback: () => void | Promise<void>
}

export default function ActionButtonsReact(props: ActionButtonsProps) {
  const { shareRows, opening, onOpenRow, onCopyLink, onCopyPage, onFeedback } = props

  return (
    <Space wrap spacing={8}>
      {shareRows.map((row, idx) => (
        <Button
          key={`${idx}-${row.url.slice(0, 64)}`}
          theme="solid"
          type="primary"
          loading={opening}
          disabled={opening || shareRows.length === 0}
          onClick={() => void onOpenRow(row.url)}
        >
          {row.label}
        </Button>
      ))}
      <Button theme="outline" onClick={() => void onCopyLink()}>
        复制资源链接
      </Button>
      <Button theme="outline" onClick={() => void onCopyPage()}>
        复制页面地址
      </Button>
      <Button theme="light" type="danger" onClick={() => void onFeedback()}>
        反馈问题
      </Button>
    </Space>
  )
}
