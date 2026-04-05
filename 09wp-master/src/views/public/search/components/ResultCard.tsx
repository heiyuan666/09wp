import { type MouseEvent, useState } from 'react'
import { Button, Card, Modal, Select, Tag, TextArea, Toast, Typography } from '@douyinfe/semi-ui'
import { feedbackCreate } from '@/api/feedback'
import {
  categoryText,
  driveIconSrc,
  extractFiles,
  formatDate,
  platformText,
  tagsOf,
  type ICategory,
  type ISearchResource,
} from '../searchHelpers'
import styles from '../styles/SearchPage.module.scss'

const { Text } = Typography

const feedbackTypeOptions = [
  { value: 'report_feedback', label: '举报反馈' },
  { value: 'content_error', label: '内容异常' },
  { value: 'password_error', label: '密码错误' },
  { value: 'other', label: '其他问题' },
]

function HighlightText({ text, keyword }: { text: string; keyword: string }) {
  if (!keyword.trim()) return <>{text}</>
  const escapedKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const parts = text.split(new RegExp(`(${escapedKeyword})`, 'gi'))
  return (
    <>
      {parts.map((part, index) => {
        const matched = part.toLowerCase() === keyword.toLowerCase()
        return (
          <span key={`${part}-${index}`} className={matched ? styles.highlight : ''}>
            {part}
          </span>
        )
      })}
    </>
  )
}

type ResultCardProps = {
  item: ISearchResource
  keyword: string
  categories: ICategory[]
  onGoDetail: (id: string | number) => void
}

export default function ResultCard({ item, keyword, categories, onGoDetail }: ResultCardProps) {
  const files = extractFiles(item)
  const tags = tagsOf(item.tags)
  const [reportingInvalid, setReportingInvalid] = useState(false)
  const [feedbackVisible, setFeedbackVisible] = useState(false)
  const [feedbackType, setFeedbackType] = useState('report_feedback')
  const [feedbackText, setFeedbackText] = useState('')
  const [reportingFeedback, setReportingFeedback] = useState(false)

  const numericResourceID = Number(item.id)
  const canReport = Number.isFinite(numericResourceID) && numericResourceID > 0

  const stopEvent = (event: MouseEvent) => {
    event.preventDefault()
    event.stopPropagation()
  }

  const submitInvalidReport = async () => {
    if (!canReport) {
      Toast.error('资源 ID 无效，无法提交举报')
      return
    }
    setReportingInvalid(true)
    try {
      const { data: res } = await feedbackCreate({
        resource_id: numericResourceID,
        type: 'link_invalid',
        content: `搜索页失效举报：${item.title}`,
      })
      if (res.code !== 200) return
      Toast.success('失效举报已提交，感谢反馈')
    } finally {
      setReportingInvalid(false)
    }
  }

  const openFeedbackModal = (event: MouseEvent) => {
    stopEvent(event)
    setFeedbackType('report_feedback')
    setFeedbackText('')
    setFeedbackVisible(true)
  }

  const submitFeedback = async () => {
    if (!canReport) {
      Toast.error('资源 ID 无效，无法提交反馈')
      return
    }
    const content = feedbackText.trim()
    if (!content) {
      Toast.warning('请先填写反馈内容')
      return
    }

    setReportingFeedback(true)
    try {
      const { data: res } = await feedbackCreate({
        resource_id: numericResourceID,
        type: feedbackType,
        content,
      })
      if (res.code !== 200) return
      Toast.success('反馈已提交，我们会尽快处理')
      setFeedbackVisible(false)
      setFeedbackText('')
    } finally {
      setReportingFeedback(false)
    }
  }

  return (
    <>
      <Card className={styles.resultCard} shadows="hover" bodyStyle={{ padding: 16 }}>
        <div className={styles.resultTitleRow}>
          <div className={styles.resultTitleLeft} onClick={() => onGoDetail(item.id)} role="button" tabIndex={0}>
            <img className={styles.resultIcon} src={driveIconSrc(item.link)} alt="" />
            <Text strong className={styles.resultTitleText}>
              <HighlightText text={item.title} keyword={keyword} />
            </Text>
          </div>
        </div>

        {files.length > 0 ? (
          <div className={styles.fileList}>
            {files.slice(0, 3).map((file) => (
              <div key={file.name} className={styles.fileLine}>
                <span className={styles.filePrefix}>file:</span>
                <span>
                  <HighlightText text={file.name} keyword={keyword} />
                </span>
              </div>
            ))}
          </div>
        ) : null}

        <div className={styles.tagLine}>
          <Tag color="grey">{categoryText(categories, item.category_id)}</Tag>
          {tags.slice(0, 5).map((tag) => (
            <Tag key={`${item.id}-${tag}`} color="grey">
              <HighlightText text={tag} keyword={keyword} />
            </Tag>
          ))}
        </div>

        <div className={styles.metaLine}>
          <span>网盘来源 {platformText(item.link)}</span>
          <span>日期 {formatDate(item.created_at)}</span>
        </div>

        <div
          className={styles.actionLine}
          onClick={(event) => event.stopPropagation()}
          onMouseDown={(event) => event.stopPropagation()}
        >
          <Button
            type="tertiary"
            theme="borderless"
            onClick={(event) => {
              stopEvent(event)
              onGoDetail(item.id)
            }}
          >
            查看
          </Button>
          <Button
            type="tertiary"
            theme="borderless"
            loading={reportingInvalid}
            onClick={(event) => {
              stopEvent(event)
              void submitInvalidReport()
            }}
          >
            失效举报
          </Button>
          <Button type="tertiary" theme="borderless" onClick={openFeedbackModal}>
            举报反馈
          </Button>
        </div>
      </Card>

      <Modal
        title="举报反馈"
        visible={feedbackVisible}
        onCancel={() => {
          if (reportingFeedback) return
          setFeedbackVisible(false)
        }}
        onOk={() => void submitFeedback()}
        okText="提交反馈"
        cancelText="取消"
        confirmLoading={reportingFeedback}
      >
        <div style={{ display: 'grid', gap: 12 }}>
          <Select
            value={feedbackType}
            optionList={feedbackTypeOptions}
            onChange={(value) => setFeedbackType(String(value))}
            style={{ width: '100%' }}
          />
          <TextArea
            value={feedbackText}
            onChange={(value) => setFeedbackText(String(value))}
            placeholder="请填写举报原因或补充信息"
            maxCount={500}
            rows={5}
            showClear
          />
        </div>
      </Modal>
    </>
  )
}
