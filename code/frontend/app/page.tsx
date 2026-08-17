import ViewTasksByStatus from "../components/ViewTasksByStatus";

export default function Home() {
  return (
    <main className="page-shell">
      <section className="composition-slot" aria-label="Task board application">
        <ViewTasksByStatus />
      </section>
    </main>
  );
}
