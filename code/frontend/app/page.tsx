import CreateTaskPanel from "../components/CreateTaskPanel";

export default function Home() {
  return (
    <main className="page-shell">
      <section className="composition-slot" aria-label="Task board application">
        <CreateTaskPanel />
        {/* Story components mount here. */}
      </section>
    </main>
  );
}
