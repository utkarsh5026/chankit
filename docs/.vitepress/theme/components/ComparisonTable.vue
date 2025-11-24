<template>
  <div class="comparison-table my-8 overflow-x-auto">
    <table class="w-full border-collapse">
      <thead>
        <tr class="bg-dark-bg-alt">
          <th class="border border-dark-border px-4 py-3 text-left font-semibold text-primary-400">Aspect</th>
          <th class="border border-dark-border px-4 py-3 text-left font-semibold text-red-400">Traditional Go</th>
          <th class="border border-dark-border px-4 py-3 text-left font-semibold text-green-400">chankit</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td class="border border-dark-border px-4 py-3 font-medium">Code Length</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">40-60 lines typical</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">8-12 lines average</td>
        </tr>
        <tr class="bg-dark-bg-soft">
          <td class="border border-dark-border px-4 py-3 font-medium">Readability</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Nested, imperative</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Declarative, fluent</td>
        </tr>
        <tr>
          <td class="border border-dark-border px-4 py-3 font-medium">Error Prone</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Channel leaks, missing close()</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Automatic cleanup</td>
        </tr>
        <tr class="bg-dark-bg-soft">
          <td class="border border-dark-border px-4 py-3 font-medium">Boilerplate</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Heavy goroutine management</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">None - handled automatically</td>
        </tr>
        <tr>
          <td class="border border-dark-border px-4 py-3 font-medium">Composability</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Difficult to combine operations</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Natural method chaining</td>
        </tr>
        <tr class="bg-dark-bg-soft">
          <td class="border border-dark-border px-4 py-3 font-medium">Testing</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Complex setup required</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Easy to test pure functions</td>
        </tr>
        <tr>
          <td class="border border-dark-border px-4 py-3 font-medium">Maintenance</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Hard to modify pipelines</td>
          <td class="border border-dark-border px-4 py-3 text-dark-text-muted">Add/remove operators easily</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.comparison-table table {
  font-size: 0.95rem;
}

.comparison-table th {
  text-transform: uppercase;
  font-size: 0.85rem;
  letter-spacing: 0.05em;
}

.comparison-table tr {
  transition: background-color 0.2s;
}

.comparison-table tbody tr:hover {
  background-color: rgba(14, 165, 233, 0.05);
}
</style>
