import { IconAlertCircle, IconCopy, IconFolderOpen, IconLink } from '@douyinfe/semi-icons'
import { Button } from '@douyinfe/semi-ui'

type ActionButtonsProps = {
  canOpen: boolean
  opening: boolean
  onCopyLink: () => void
  onOpen: () => void
  onCopyPage: () => void
  onFeedback: () => void
}

export default function ActionButtonsReact(props: ActionButtonsProps) {
  const buttonClass =
    'group !h-10 !rounded-xl !px-3 !text-[14px] !font-medium !text-slate-700 hover:!text-sky-600 hover:!bg-sky-50/90 hover:!shadow-sm active:!scale-[0.98] transition-all duration-200'
  const iconClass = 'transition-transform duration-200 group-hover:scale-110 group-hover:-translate-y-0.5'

  return (
    <div className="flex flex-wrap items-center gap-1.5 rounded-2xl border border-slate-100/80 bg-white/90 px-2 py-1.5 shadow-[0_8px_24px_rgba(15,23,42,0.05)] backdrop-blur md:gap-2">
      <Button
        icon={<IconLink className={iconClass} />}
        theme="borderless"
        className={buttonClass}
        onClick={props.onCopyLink}
      >
        复制链接
      </Button>
      <Button
        icon={<IconFolderOpen className={iconClass} />}
        theme="borderless"
        className={buttonClass}
        disabled={!props.canOpen}
        loading={props.opening}
        onClick={props.onOpen}
      >
        查看资源
      </Button>
      <Button
        icon={<IconCopy className={iconClass} />}
        theme="borderless"
        className={buttonClass}
        onClick={props.onCopyPage}
      >
        复制本页
      </Button>
      <Button
        icon={<IconAlertCircle className={iconClass} />}
        theme="borderless"
        className={buttonClass}
        onClick={props.onFeedback}
      >
        反馈问题
      </Button>
    </div>
  )
}
