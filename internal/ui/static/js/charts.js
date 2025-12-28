function allocationChart() {
    return {
        chart: null,
        init() {
            this.$nextTick(() => {
                const ctx = this.$refs.canvas.getContext('2d');
                const dataEl = this.$refs.canvas.closest('[data-chart-items]');
                if (!dataEl) return;
                const items = JSON.parse(dataEl.dataset.chartItems);

                var existingChart = Chart.getChart(ctx);
                if (existingChart) {
                    existingChart.destroy();
                }

                this.chart = new Chart(ctx, {
                    type: 'doughnut',
                    data: {
                        labels: items.map(i => i.label),
                        datasets: [{
                            data: items.map(i => i.value),
                            backgroundColor: items.map(i => i.color),
                            borderWidth: 0,
                            hoverOffset: 4
                        }]
                    },
                    options: {
                        responsive: false,
                        maintainAspectRatio: false,
                        cutout: '70%',
                        plugins: {
                            legend: { display: false },
                            tooltip: {
                                backgroundColor: '#1a1f2e',
                                titleColor: '#fff',
                                bodyColor: '#8892a0',
                                borderColor: '#2d3548',
                                borderWidth: 1,
                                padding: 12,
                                callbacks: {
                                    label: function(ctx) {
                                        return ctx.label + ': ' + ctx.parsed.toFixed(1) + '%';
                                    }
                                }
                            }
                        }
                    }
                });
            });
        }
    }
}

function createLineChart(ctx, labels, values, maxTicksLimit) {
    var existingChart = Chart.getChart(ctx);
    if (existingChart) {
        existingChart.destroy();
    }
    return new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Portfolio Value',
                data: values,
                borderColor: '#f7931a',
                backgroundColor: 'rgba(247, 147, 26, 0.1)',
                fill: true,
                tension: 0.4,
                pointRadius: 0,
                pointHoverRadius: 6,
                pointHoverBackgroundColor: '#f7931a'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            scales: {
                x: {
                    grid: { display: false },
                    ticks: {
                        maxTicksLimit: maxTicksLimit,
                        color: '#8892a0'
                    }
                },
                y: {
                    grid: { color: 'rgba(255,255,255,0.05)' },
                    ticks: {
                        color: '#8892a0',
                        callback: function(v) {
                            if (v >= 1000000) return '$' + (v/1000000).toFixed(1) + 'M';
                            if (v >= 1000) return '$' + (v/1000).toFixed(1) + 'K';
                            return '$' + v.toFixed(0);
                        }
                    }
                }
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: '#1a1f2e',
                    titleColor: '#fff',
                    bodyColor: '#8892a0',
                    borderColor: '#2d3548',
                    borderWidth: 1,
                    padding: 12,
                    displayColors: false,
                    callbacks: {
                        label: function(ctx) {
                            return '$' + ctx.parsed.y.toLocaleString(undefined, {minimumFractionDigits: 2, maximumFractionDigits: 2});
                        }
                    }
                }
            }
        }
    });
}

function portfolioChart() {
    return {
        chart: null,
        init() {
            this.$nextTick(() => {
                const ctx = this.$refs.canvas.getContext('2d');
                const dataEl = this.$refs.canvas.closest('[data-chart-labels]');
                if (!dataEl) return;
                const labels = JSON.parse(dataEl.dataset.chartLabels);
                const values = JSON.parse(dataEl.dataset.chartValues);
                this.chart = createLineChart(ctx, labels, values, 6);
            });
        }
    }
}

function portfolioHistoryChart() {
    return {
        chart: null,
        init() {
            this.$nextTick(() => {
                const ctx = this.$refs.canvas.getContext('2d');
                const dataEl = this.$refs.canvas.closest('[data-chart-labels]');
                if (!dataEl) return;
                const labels = JSON.parse(dataEl.dataset.chartLabels);
                const values = JSON.parse(dataEl.dataset.chartValues);
                this.chart = createLineChart(ctx, labels, values, 8);
            });
        }
    }
}
