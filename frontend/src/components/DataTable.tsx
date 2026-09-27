import { flexRender, getCoreRowModel, useReactTable, type ColumnDef, type RowData } from '@tanstack/react-table'

declare module '@tanstack/react-table' {
  // TData и TValue совпадают с объявлением в библиотеке, сами поля их не читают.
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData extends RowData, TValue> {
    align?: 'right'
    className?: string
  }
}

type DataTableProps<T> = {
  columns: ColumnDef<T>[]
  data: T[]
  getRowId?: (row: T, index: number) => string
  minWidthClass?: string
}

function cellClass(meta: { align?: 'right'; className?: string } | undefined): string {
  return ['px-3 py-2 align-middle', meta?.align === 'right' ? 'text-right' : '', meta?.className ?? '']
    .filter(Boolean)
    .join(' ')
}

/** Таблица в том же виде, что список записей: шапка, строки, горизонтальный скролл на узком экране. */
export function DataTable<T>({ columns, data, getRowId, minWidthClass = 'min-w-[36rem]' }: DataTableProps<T>) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getRowId,
  })

  return (
    <div className="overflow-x-auto">
      <table className={`w-full text-left text-sm ${minWidthClass}`}>
        <thead className="bg-slate-900 text-xs tracking-wide text-slate-400 uppercase">
          {table.getHeaderGroups().map((group) => (
            <tr key={group.id}>
              {group.headers.map((header) => (
                <th key={header.id} className={`font-medium ${cellClass(header.column.columnDef.meta)}`}>
                  {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id} className="border-t border-slate-800 hover:bg-slate-900/60">
              {row.getVisibleCells().map((cell) => (
                <td key={cell.id} className={cellClass(cell.column.columnDef.meta)}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
