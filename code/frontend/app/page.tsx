import DeleteTask from "../components/DeleteTask";

export default function Home() {
  return (
    <main className="page-shell">
      <section className="composition-slot" aria-label="Task board application">
        {/* Story components mount here. */}
        <DeleteTask />
      </section>
    </main>
  );
}
