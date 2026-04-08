export function JsonLd({ data }: { data: unknown }) {
  return (
    <script
      type="application/ld+json"
      // json-ld must be raw JSON string
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  )
}

