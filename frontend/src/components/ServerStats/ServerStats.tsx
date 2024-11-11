import React, { useState, useEffect } from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { Calendar, Search, Layout, Clock } from "lucide-react";
import styles from "./ServerStats.module.css";

interface ServerStats {
  schedule_requests: number;
  total_schedules: number;
  search_requests: number;
  subject_counts: Record<string, number>;
  last_course_update: string;
  server_creation_date: string;
  avg_schedules_per_request: number;
}

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
}

const StatCard: React.FC<StatCardProps> = ({ title, value, icon }) => (
  <div className={styles.statCard}>
    <div className={styles.iconContainer}>{icon}</div>
    <div className={styles.statContent}>
      <div className={styles.statLabel}>{title}</div>
      <div className={styles.statValue}>{value}</div>
    </div>
  </div>
);

const ServerStats: React.FC = () => {
  const [stats, setStats] = useState<ServerStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [excludedSubjects, setExcludedSubjects] = useState<Set<string>>(
    new Set()
  );

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch("/schedule-optimizer/stats");
        if (!response.ok) {
          throw new Error("Failed to fetch stats");
        }
        const data = await response.json();
        console.log("Stats: ", data);
        setStats(data);
      } catch (err) {
        setError("Failed to load statistics");
        console.error("Error fetching stats:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  const handleBarClick = (data: any) => {
    setExcludedSubjects((prev) => {
      const newExcluded = new Set(prev);
      newExcluded.add(data.subject);
      return newExcluded;
    });
  };

  const handleResetFilters = () => {
    setExcludedSubjects(new Set());
  };

  if (loading) {
    return <div className={styles.loading}>Loading statistics...</div>;
  }

  if (error || !stats) {
    return (
      <div className={styles.error}>{error || "Failed to load statistics"}</div>
    );
  }

  // Prepare data for the bar chart
  const subjectData = Object.entries(stats.subject_counts)
    .map(([subject, count]) => ({
      subject,
      count,
    }))
    .filter((item) => !excludedSubjects.has(item.subject))
    .sort((a, b) => b.count - a.count)
    .slice(0, 10); // Show top 15 subjects

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  return (
    <div className={styles.container}>
      <h2 className={styles.title}>Server Statistics</h2>

      <div className={styles.statsGrid}>
        <StatCard
          title="Schedule Requests"
          value={stats.schedule_requests.toLocaleString()}
          icon={<Calendar className={styles.icon} size={24} />}
        />
        <StatCard
          title="Search Requests"
          value={stats.search_requests.toLocaleString()}
          icon={<Search className={styles.icon} size={24} />}
        />
        <StatCard
          title="Total Schedules Generated"
          value={stats.total_schedules.toLocaleString()}
          icon={<Layout className={styles.icon} size={24} />}
        />
        <StatCard
          title="Avg. Schedules/Request"
          value={stats.avg_schedules_per_request.toFixed(1)}
          icon={<Clock className={styles.icon} size={24} />}
        />
      </div>

      <div className={styles.chartContainer}>
        <div className={styles.chartHeader}>
          <h3 className={styles.chartTitle}>
            Top 10 Subjects by Request Count
          </h3>
          {excludedSubjects.size > 0 && (
            <button className={styles.resetButton} onClick={handleResetFilters}>
              Reset Filters
            </button>
          )}
        </div>
        <div className={styles.chartSubtitle}>
          Press a subject to remove it from the chart.
        </div>
        <div className={styles.chart}>
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={subjectData}
              margin={{ top: 20, right: 30, left: 20, bottom: 60 }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey="subject"
                angle={-45}
                textAnchor="end"
                height={60}
                interval={0}
              />
              <YAxis />
              <Tooltip />
              <Bar
                dataKey="count"
                fill="var(--color-primary-600)"
                onClick={handleBarClick}
                cursor="pointer"
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className={styles.serverInfo}>
        <h3 className={styles.serverInfoTitle}>Server Information</h3>
        <div className={styles.serverInfoItem}>
          <span className={styles.serverInfoLabel}>Server Created: </span>
          {formatDate(stats.server_creation_date)}
        </div>
        <div className={styles.serverInfoItem}>
          <span className={styles.serverInfoLabel}>Last Course Update: </span>
          {stats.last_course_update === "0001-01-01T00:00:00Z"
            ? "No updates yet"
            : formatDate(stats.last_course_update)}
        </div>
      </div>
    </div>
  );
};

export default ServerStats;
